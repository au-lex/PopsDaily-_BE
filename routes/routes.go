package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Register(app *fiber.App, db *gorm.DB) {
	api := app.Group("/api")

	RegisterFeedRoutes(api, db)
	RegisterArticleRoutes(api, db)
	RegisterBookmarkRoutes(api, db)
	RegisterAuthRoutes(api, db)
	RegisterUserRoutes(api, db)
	RegisterSearchRoutes(api, db)
	// RegisterDeviceRoutes(api, db)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}