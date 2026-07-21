package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"rss-backend/models"
)

type BookmarkHandler struct {
	DB *gorm.DB
}

func NewBookmarkHandler(db *gorm.DB) *BookmarkHandler {
	return &BookmarkHandler{DB: db}
}

// POST /api/bookmarks
// body: { "device_id": "...", "article_id": 123 }
func (h *BookmarkHandler) AddBookmark(c *fiber.Ctx) error {
	var req models.CreateBookmarkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.DeviceID == "" || req.ArticleID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "device_id and article_id are required"})
	}

	// Confirm the article actually exists before bookmarking it
	var article models.Article
	if err := h.DB.First(&article, req.ArticleID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "article not found"})
	}

	bookmark := models.Bookmark{
		DeviceID:  req.DeviceID,
		ArticleID: req.ArticleID,
	}

	// If it already exists (device already bookmarked this article),
	// just return the existing one instead of erroring.
	if err := h.DB.Where("device_id = ? AND article_id = ?", req.DeviceID, req.ArticleID).
		FirstOrCreate(&bookmark).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save bookmark"})
	}

	return c.Status(fiber.StatusCreated).JSON(bookmark)
}

// DELETE /api/bookmarks/:article_id?device_id=xxx
func (h *BookmarkHandler) RemoveBookmark(c *fiber.Ctx) error {
	articleID := c.Params("article_id")
	deviceID := c.Query("device_id")

	if deviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "device_id query param is required"})
	}

	result := h.DB.Where("device_id = ? AND article_id = ?", deviceID, articleID).
		Delete(&models.Bookmark{})

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to remove bookmark"})
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "bookmark not found"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GET /api/bookmarks?device_id=xxx&page=1&limit=20
func (h *BookmarkHandler) ListBookmarks(c *fiber.Ctx) error {
	deviceID := c.Query("device_id")
	if deviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "device_id query param is required"})
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

	var total int64
	h.DB.Model(&models.Bookmark{}).Where("device_id = ?", deviceID).Count(&total)

	var bookmarks []models.Bookmark
	if err := h.DB.
		Where("device_id = ?", deviceID).
		Preload("Article").
		Preload("Article.Feed").
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&bookmarks).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch bookmarks"})
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return c.JSON(fiber.Map{
		"data":        bookmarks,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
	})
}

// GET /api/bookmarks/check/:article_id?device_id=xxx
// Handy for the article screen to know whether to show a filled/outline icon.
func (h *BookmarkHandler) CheckBookmark(c *fiber.Ctx) error {
	articleID := c.Params("article_id")
	deviceID := c.Query("device_id")

	if deviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "device_id query param is required"})
	}

	var count int64
	h.DB.Model(&models.Bookmark{}).
		Where("device_id = ? AND article_id = ?", deviceID, articleID).
		Count(&count)

	return c.JSON(fiber.Map{"bookmarked": count > 0})
}