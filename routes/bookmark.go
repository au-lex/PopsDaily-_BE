package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"rss-backend/handlers"
)

// RegisterBookmarkRoutes wires up all /api/bookmarks endpoints.
// Called from Register() in routes.go, passing in the same api group.
func RegisterBookmarkRoutes(api fiber.Router, db *gorm.DB) {
	bookmarkHandler := handlers.NewBookmarkHandler(db)

	bookmarks := api.Group("/bookmarks")
	bookmarks.Post("/", bookmarkHandler.AddBookmark)
	bookmarks.Get("/", bookmarkHandler.ListBookmarks)
	bookmarks.Delete("/:article_id", bookmarkHandler.RemoveBookmark)
	bookmarks.Get("/check/:article_id", bookmarkHandler.CheckBookmark)
}