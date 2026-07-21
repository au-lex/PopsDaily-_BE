package models

import "time"

// PasswordResetOTP stores a one-time code sent to a user's email for
// resetting their password. Old codes for the same email are invalidated
// whenever a new one is generated (see AuthService.GenerateOTP).
type PasswordResetOTP struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"index;not null" json:"email"`
	Code      string    `gorm:"not null" json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
	CreatedAt time.Time `json:"created_at"`
}