package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"rss-backend/handlers"
)

func Register(app *fiber.App, db *gorm.DB) {
	feedHandler := handlers.NewFeedHandler(db)
	articleHandler := handlers.NewArticleHandler(db)

	api := app.Group("/api")

	feeds := api.Group("/feeds")
	feeds.Post("/", feedHandler.CreateFeed)
	feeds.Get("/", feedHandler.ListFeeds)
	feeds.Get("/:id", feedHandler.GetFeed)
	feeds.Patch("/:id", feedHandler.UpdateFeed)
	feeds.Delete("/:id", feedHandler.DeleteFeed)
	feeds.Post("/:id/fetch", feedHandler.FetchFeedNow)
	feeds.Post("/fetch-all", feedHandler.FetchAllNow)

	articles := api.Group("/articles")
	articles.Get("/", articleHandler.ListArticles)
	articles.Get("/:id", articleHandler.GetArticle)

	RegisterBookmarkRoutes(api, db)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}