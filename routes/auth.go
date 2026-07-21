package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"popsdaily/handlers"
)

// RegisterAuthRoutes wires up all /api/auth endpoints: signup, login,
// forgot password, resend OTP, reset password.
func RegisterAuthRoutes(api fiber.Router, db *gorm.DB) {
	authHandler := handlers.NewAuthHandler(db)

	auth := api.Group("/auth")
	auth.Post("/signup", authHandler.Signup)
	auth.Post("/login", authHandler.Login)
	auth.Post("/forgot-password", authHandler.ForgotPassword)
	auth.Post("/resend-otp", authHandler.ResendOTP)
	auth.Post("/reset-password", authHandler.ResetPassword)
}