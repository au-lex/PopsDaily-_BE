package models

import "time"

// DeviceToken stores a Firebase Cloud Messaging token for a device, so the
// backend knows where to push notifications when new articles come in.
// A device re-registers (upserts) its token on every app launch, since FCM
// tokens can change/rotate.
type DeviceToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DeviceID  string    `gorm:"uniqueIndex;not null" json:"device_id"`
	Token     string    `gorm:"not null" json:"token"`
	Platform  string    `json:"platform"` // "android" or "ios"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterDeviceRequest struct {
	DeviceID string `json:"device_id"`
	Token    string `json:"token"`
	Platform string `json:"platform"`
}