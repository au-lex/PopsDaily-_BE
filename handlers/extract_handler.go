// handlers/extract_handler.go
package handlers

import (
	"net/http"
	"net/url"
	"time"

	readability "github.com/go-shiori/go-readability"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"popsdaily/models"
)

type ExtractHandler struct {
	DB *gorm.DB
}

func NewExtractHandler(db *gorm.DB) *ExtractHandler {
	return &ExtractHandler{DB: db}
}

// GET /api/articles/:id/extract
func (h *ExtractHandler) ExtractContent(c *fiber.Ctx) error {
	id := c.Params("id")

	var article models.Article
	if err := h.DB.First(&article, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "article not found"})
	}

	// If we already have real content, no need to scrape.
	if len(article.Content) > 500 {
		return c.JSON(fiber.Map{
			"title":   article.Title,
			"content": article.Content,
			"cached":  false,
			"source":  "db",
		})
	}

	parsedURL, err := url.Parse(article.Link)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid article link"})
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(article.Link)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to fetch article page"})
	}
	defer resp.Body.Close()

	result, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "could not extract readable content"})
	}

	if result.TextContent == "" {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "no readable content found"})
	}

	// Optionally cache it back onto the article row so future requests skip scraping.
	h.DB.Model(&article).Update("content", result.Content)

	return c.JSON(fiber.Map{
		"title":   result.Title,
		"content": result.Content,
		"cached":  true,
		"source":  "extracted",
	})
}