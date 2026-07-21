package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"popsdaily/models"
)

var DB *gorm.DB

// Connect opens a Postgres connection using a standard GORM DSN, e.g:
// "host=localhost user=postgres password=postgres dbname=rss_backend port=5432 sslmode=disable"
func Connect(dsn string) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&models.Feed{}, &models.Article{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	DB = db
	log.Println("database connected and migrated")
}
