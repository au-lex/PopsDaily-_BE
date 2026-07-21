package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"popsdaily/database"
	"popsdaily/routes"
	"popsdaily/services"
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

if err := database.Migrate(); err != nil {
	log.Fatalf("migration failed: %v", err)
}

	if err := services.InitFirebase(); err != nil {
		log.Printf("firebase init warning: %v", err)
	}

	pollSpec := os.Getenv("POLL_INTERVAL")
	if pollSpec == "" {
		pollSpec = "@every 15m"
	}
	services.StartPoller(database.DB, pollSpec)

	app := fiber.New(fiber.Config{
		AppName: "popsdaily",
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