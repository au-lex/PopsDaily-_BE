package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"popsdaily/models"
	"popsdaily/services"
)

type FeedHandler struct {
	DB *gorm.DB
}

func NewFeedHandler(db *gorm.DB) *FeedHandler {
	return &FeedHandler{DB: db}
}

// POST /api/feeds
func (h *FeedHandler) CreateFeed(c *fiber.Ctx) error {
	var req models.CreateFeedRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Name == "" || req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name and url are required"})
	}

	feed := models.Feed{
		Name:     req.Name,
		URL:      req.URL,
		Category: req.Category,
		Active:   true,
	}

	if err := h.DB.Create(&feed).Error; err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "feed already exists or could not be saved: " + err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(feed)
}

// GET /api/feeds
func (h *FeedHandler) ListFeeds(c *fiber.Ctx) error {
	var feeds []models.Feed
	query := h.DB.Model(&models.Feed{})

	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if active := c.Query("active"); active != "" {
		query = query.Where("active = ?", active == "true")
	}

	if err := query.Order("created_at desc").Find(&feeds).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch feeds"})
	}
	return c.JSON(feeds)
}

// GET /api/feeds/:id
func (h *FeedHandler) GetFeed(c *fiber.Ctx) error {
	id := c.Params("id")
	var feed models.Feed
	if err := h.DB.First(&feed, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "feed not found"})
	}
	return c.JSON(feed)
}

// PATCH /api/feeds/:id
func (h *FeedHandler) UpdateFeed(c *fiber.Ctx) error {
	id := c.Params("id")
	var feed models.Feed
	if err := h.DB.First(&feed, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "feed not found"})
	}

	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	// Only allow specific fields to be patched
	allowed := map[string]bool{"name": true, "url": true, "category": true, "active": true}
	safeUpdates := map[string]interface{}{}
	for k, v := range updates {
		if allowed[k] {
			safeUpdates[k] = v
		}
	}

	if err := h.DB.Model(&feed).Updates(safeUpdates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update feed"})
	}
	return c.JSON(feed)
}

// DELETE /api/feeds/:id
func (h *FeedHandler) DeleteFeed(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.DB.Delete(&models.Feed{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete feed"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// POST /api/feeds/:id/fetch
func (h *FeedHandler) FetchFeedNow(c *fiber.Ctx) error {
	id := c.Params("id")
	var feed models.Feed
	if err := h.DB.First(&feed, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "feed not found"})
	}

	inserted, err := services.FetchAndStoreFeed(h.DB, &feed)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to fetch feed: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"feed_id":       feed.ID,
		"new_articles":  inserted,
		"fetched_at_ok": true,
	})
}

// POST /api/feeds/fetch-all
func (h *FeedHandler) FetchAllNow(c *fiber.Ctx) error {
	go services.PollAllFeeds(h.DB) // run in background, respond immediately
	return c.JSON(fiber.Map{"message": "poll triggered for all active feeds"})
}

func atoi(s string, fallback int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}
