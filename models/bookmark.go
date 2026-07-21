package models

import "time"

// Bookmark links a user to a saved article. Since this app currently has
// no auth system, DeviceID is used to scope bookmarks per installation
// (e.g. a UUID generated once and stored locally in the Flutter app).
// If you add real user accounts later, swap DeviceID for UserID.
type Bookmark struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	DeviceID  string  `gorm:"uniqueIndex:idx_device_article;not null" json:"device_id"`
	ArticleID uint    `gorm:"uniqueIndex:idx_device_article;not null" json:"article_id"`
	Article   Article `gorm:"foreignKey:ArticleID" json:"article,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateBookmarkRequest struct {
	DeviceID  string `json:"device_id" validate:"required"`
	ArticleID uint   `json:"article_id" validate:"required"`
}