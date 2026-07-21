package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"popsdaily/handlers"
)


func RegisterArticleRoutes(api fiber.Router, db *gorm.DB) {
	articleHandler := handlers.NewArticleHandler(db)

	articles := api.Group("/articles")
	articles.Get("/", articleHandler.ListArticles)
	articles.Get("/:id", articleHandler.GetArticle)
}