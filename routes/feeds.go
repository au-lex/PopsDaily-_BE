package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"popsdaily/handlers"
)


func RegisterFeedRoutes(api fiber.Router, db *gorm.DB) {
	feedHandler := handlers.NewFeedHandler(db)

	feeds := api.Group("/feeds")
	feeds.Post("/", feedHandler.CreateFeed)
	feeds.Get("/", feedHandler.ListFeeds)
	feeds.Get("/sources", feedHandler.ListSources)  
	feeds.Get("/:id", feedHandler.GetFeed)
	feeds.Patch("/:id", feedHandler.UpdateFeed)
	feeds.Delete("/:id", feedHandler.DeleteFeed)
	feeds.Post("/:id/fetch", feedHandler.FetchFeedNow)
	feeds.Post("/fetch-all", feedHandler.FetchAllNow)
}