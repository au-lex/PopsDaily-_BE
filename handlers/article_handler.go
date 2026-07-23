package handlers

import (
	"strconv"

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