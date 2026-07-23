package models

import "time"

type Feed struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Name          string     `gorm:"not null" json:"name"`
	Source        string     `gorm:"index" json:"source"` // temporarily no "not null"
	URL           string     `gorm:"uniqueIndex;not null" json:"url"`
	Category      string     `json:"category"`
	Active        bool       `gorm:"default:true" json:"active"`
	LastFetchedAt *time.Time `json:"last_fetched_at"`
	LastError     string     `json:"last_error,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CreateFeedRequest struct {
	Name     string `json:"name" validate:"required"`
	Source   string `json:"source" validate:"required"` // NEW
	URL      string `json:"url" validate:"required"`
	Category string `json:"category"`
}