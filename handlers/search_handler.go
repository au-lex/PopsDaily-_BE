package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"rss-backend/models"
)

type SearchHandler struct {
	DB *gorm.DB
}

func NewSearchHandler(db *gorm.DB) *SearchHandler {
	return &SearchHandler{DB: db}
}

// GET /api/search?q=tinubu&tag=politics&category=national&page=1&limit=20
//
// All params are optional and combine with AND:
//   q        — free text search across title, description, content
//   tag      — matches articles whose tags list contains this tag
//   category — exact category match (national, politics, sports, etc.)
//   feed_id  — restrict to one specific feed/source
func (h *SearchHandler) Search(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q"))
	tag := strings.TrimSpace(c.Query("tag"))
	category := strings.TrimSpace(c.Query("category"))
	feedID := c.Query("feed_id")

	page := atoi(c.Query("page", "1"), 1)
	limit := atoi(c.Query("limit", "20"), 20)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	if q == "" && tag == "" && category == "" && feedID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provide at least one of: q, tag, category, feed_id",
		})
	}

	query := h.DB.Model(&models.Article{})

	if q != "" {
		like := "%" + q + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ? OR content ILIKE ?", like, like, like)
	}
	if tag != "" {
		// Tags are stored comma-separated (e.g. "politics,tinubu,senate"),
		// so match it as a substring bounded by commas or string edges.
		query = query.Where("tags ILIKE ?", "%"+tag+"%")
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if feedID != "" {
		query = query.Where("feed_id = ?", feedID)
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "search failed"})
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

// GET /api/search/tags
// Returns every distinct tag currently in use, handy for building a
// "browse by tag" chip list or autocomplete in the Flutter app.
func (h *SearchHandler) ListTags(c *fiber.Ctx) error {
	var rows []string
	if err := h.DB.Model(&models.Article{}).
		Where("tags != ''").
		Pluck("tags", &rows).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load tags"})
	}

	seen := map[string]bool{}
	var tags []string
	for _, row := range rows {
		for _, t := range strings.Split(row, ",") {
			t = strings.TrimSpace(t)
			if t == "" || seen[t] {
				continue
			}
			seen[t] = true
			tags = append(tags, t)
		}
	}

	return c.JSON(fiber.Map{"tags": tags, "count": len(tags)})
}