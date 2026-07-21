package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"rss-backend/handlers"
	"rss-backend/middleware"
)


func RegisterUserRoutes(api fiber.Router, db *gorm.DB) {
	authHandler := handlers.NewAuthHandler(db)

	users := api.Group("/users", middleware.RequireAuth)
	users.Get("/me", authHandler.GetProfile)
	users.Patch("/me", authHandler.UpdateProfile)
	users.Post("/change-password", authHandler.ChangePassword)
}