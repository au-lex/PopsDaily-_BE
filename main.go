package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"rss-backend/database"
	"rss-backend/routes"
	"rss-backend/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on environment variables")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set, e.g. host=localhost user=postgres password=postgres dbname=rss_backend port=5432 sslmode=disable")
	}
	database.Connect(dsn)

	pollSpec := os.Getenv("POLL_INTERVAL")
	if pollSpec == "" {
		pollSpec = "@every 15m"
	}
	services.StartPoller(database.DB, pollSpec)

	app := fiber.New(fiber.Config{
		AppName: "rss-backend",
	})

	app.Use(logger.New())
	app.Use(cors.New())

	routes.Register(app, database.DB)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}
