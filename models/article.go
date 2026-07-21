package models

import "time"

type Article struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FeedID      uint      `gorm:"index;not null" json:"feed_id"`
	Feed        Feed      `gorm:"foreignKey:FeedID" json:"feed,omitempty"`
	Title       string    `gorm:"not null" json:"title"`
	Link        string    `gorm:"index;not null" json:"link"`
	GUID        string    `gorm:"uniqueIndex;not null" json:"guid"`
	Description string    `gorm:"type:text" json:"description"`
	Content     string    `gorm:"type:text" json:"content,omitempty"`
	Author      string    `json:"author,omitempty"`
	ImageURL    string    `json:"image_url,omitempty"`
	Category    string    `json:"category,omitempty"`
	Tags        string    `gorm:"type:text;index" json:"tags,omitempty"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type ArticleListResponse struct {
	Data       []Article `json:"data"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	Total      int64     `json:"total"`
	TotalPages int       `json:"total_pages"`
}