package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"popsdaily/handlers"
)


func RegisterArticleRoutes(api fiber.Router, db *gorm.DB) {
	articleHandler := handlers.NewArticleHandler(db)
	extractHandler := handlers.NewExtractHandler(db)

	articles := api.Group("/articles")
	articles.Get("/", articleHandler.ListArticles)
	articles.Get("/source-id/:id", articleHandler.ListArticlesBySourceID)
	articles.Get("/:id/extract", extractHandler.ExtractContent)

	articles.Get("/:id", articleHandler.GetArticle)
}