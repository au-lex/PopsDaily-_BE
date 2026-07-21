package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"popsdaily/handlers"
)

// RegisterSearchRoutes wires up /api/search endpoints. Public — no auth
// required, since searching articles should work for anyone using the app.
func RegisterSearchRoutes(api fiber.Router, db *gorm.DB) {
	searchHandler := handlers.NewSearchHandler(db)

	search := api.Group("/search")
	search.Get("/", searchHandler.Search)
	search.Get("/tags", searchHandler.ListTags)
}