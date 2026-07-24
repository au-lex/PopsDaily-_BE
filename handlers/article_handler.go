package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"popsdaily/models"
)

type ArticleHandler struct {
	DB *gorm.DB
}

func NewArticleHandler(db *gorm.DB) *ArticleHandler {
	return &ArticleHandler{DB: db}
}

// GET /api/articles?page=1&limit=20&feed_id=1&category=politics&search=nigeria
func (h *ArticleHandler) ListArticles(c *fiber.Ctx) error {
	page := atoi(c.Query("page", "1"), 1)
	limit := atoi(c.Query("limit", "20"), 20)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := h.DB.Model(&models.Article{})

	if feedID := c.Query("feed_id"); feedID != "" {
		query = query.Where("feed_id = ?", feedID)
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if search := c.Query("search"); search != "" {
		like := "%" + search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}

	var total int64
	query.Count(&total)

	var articles []models.Article
	if err := query.
		Preload("Feed").
		Order("published_at desc").
		Limit(limit).
		Offset(offset).
		Find(&articles).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch articles"})
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return c.JSON(models.ArticleListResponse{
		Data:       articles,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GET /api/articles/:id
func (h *ArticleHandler) GetArticle(c *fiber.Ctx) error {
	id := c.Params("id")
	var article models.Article
	if err := h.DB.Preload("Feed").First(&article, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "article not found"})
	}
	return c.JSON(article)
}

// GET /api/articles/source-id/:id?page=1&limit=20&category=politics
func (h *ArticleHandler) ListArticlesBySourceID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid source id"})
	}

	source, ok := sourceIDMap[id]
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "unknown source id"})
	}

	page := atoi(c.Query("page", "1"), 1)
	limit := atoi(c.Query("limit", "20"), 20)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := h.DB.Model(&models.Article{}).
		Joins("JOIN feeds ON feeds.id = articles.feed_id").
		Where("feeds.source = ?", source)

	if category := c.Query("category"); category != "" {
		query = query.Where("articles.category = ?", category)
	}

	var total int64
	query.Count(&total)

	var articles []models.Article
	if err := query.
		Preload("Feed").
		Order("articles.published_at desc").
		Limit(limit).
		Offset(offset).
		Find(&articles).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch articles"})
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return c.JSON(models.ArticleListResponse{
		Data:       articles,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GET /api/articles/:id/summary
// Summaries are persisted on the article row (models.Article.Summary, a
// JSON-encoded []string) instead of an in-memory cache, so they survive
// backend restarts and work across multiple instances too.
func (h *ArticleHandler) GenerateArticleSummary(c *fiber.Ctx) error {
	articleID := c.Params("id")

	var article models.Article
	if err := h.DB.First(&article, "id = ?", articleID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "article not found"})
	}

	if article.Summary != "" {
		var cached []string
		if err := json.Unmarshal([]byte(article.Summary), &cached); err == nil {
			return c.JSON(fiber.Map{"summary": cached})
		}
		// fall through and regenerate if the stored value somehow isn't valid JSON
	}

	// Content holds the full article body; fall back to Description if
	// Content hasn't been populated (e.g. extraction hasn't run for this
	// article yet).
	textToSummarize := article.Content
	if textToSummarize == "" {
		textToSummarize = article.Description
	}

	summary, err := callOpenAISummary(article.Title, textToSummarize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "summary generation failed"})
	}

	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to encode summary"})
	}

	if err := h.DB.Model(&article).Update("summary", string(summaryJSON)).Error; err != nil {
		// generation succeeded even if the save failed — still return it,
		// just log so you notice the write is failing
		fmt.Printf("[GenerateArticleSummary] failed to persist summary for article %s: %v\n", articleID, err)
	}

	return c.JSON(fiber.Map{"summary": summary})
}

func callOpenAISummary(title, content string) ([]string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	reqBody := map[string]interface{}{
		"model": "gpt-4o-mini", // cheap, fast, plenty good for summarization
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Summarize news articles into exactly 3 short, factual bullet points. Return ONLY the bullets, one per line, no numbering, no preamble.",
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("Title: %s\n\nArticle: %s", title, content),
			},
		},
		"temperature": 0.3,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai returned status %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no completion returned")
	}

	lines := strings.Split(strings.TrimSpace(result.Choices[0].Message.Content), "\n")
	var bullets []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			bullets = append(bullets, l)
		}
	}
	return bullets, nil
}