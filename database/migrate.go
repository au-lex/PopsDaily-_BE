package database

import (
	"fmt"
	"log"

	"popsdaily/models"
)

func Migrate() error {
	log.Println("Running database migrations...")

	err := DB.AutoMigrate(
		&models.Feed{},
		&models.Article{},
		&models.Bookmark{},
		&models.User{},
		&models.PasswordResetOTP{},
		&models.DeviceToken{},
	)
	if err != nil {
		log.Printf("Error migrating database: %v", err)
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}