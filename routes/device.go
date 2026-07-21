package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"popsdaily/handlers"
)

// RegisterDeviceRoutes wires up /api/devices endpoints for FCM token
// registration, so the poller knows where to send push notifications.
func RegisterDeviceRoutes(api fiber.Router, db *gorm.DB) {
	deviceHandler := handlers.NewDeviceHandler(db)

	devices := api.Group("/devices")
	devices.Post("/", deviceHandler.RegisterDevice)
	devices.Delete("/:device_id", deviceHandler.UnregisterDevice)
}