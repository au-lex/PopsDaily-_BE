package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"popsdaily/models"
)

type DeviceHandler struct {
	DB *gorm.DB
}

func NewDeviceHandler(db *gorm.DB) *DeviceHandler {
	return &DeviceHandler{DB: db}
}

// POST /api/devices
// Registers or updates a device's FCM token. Called on app launch and
// whenever Firebase issues a new token (FCM tokens can rotate).
func (h *DeviceHandler) RegisterDevice(c *fiber.Ctx) error {
	var req models.RegisterDeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.DeviceID == "" || req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "device_id and token are required"})
	}

	var device models.DeviceToken
	err := h.DB.Where("device_id = ?", req.DeviceID).First(&device).Error

	if err == gorm.ErrRecordNotFound {
		device = models.DeviceToken{
			DeviceID: req.DeviceID,
			Token:    req.Token,
			Platform: req.Platform,
		}
		if err := h.DB.Create(&device).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to register device"})
		}
		return c.Status(fiber.StatusCreated).JSON(device)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to check existing device"})
	}

	// Device already known — update its token in case it rotated.
	h.DB.Model(&device).Updates(map[string]interface{}{
		"token":    req.Token,
		"platform": req.Platform,
	})

	return c.JSON(device)
}

// DELETE /api/devices/:device_id
// Unregister a device (e.g. user disabled notifications).
func (h *DeviceHandler) UnregisterDevice(c *fiber.Ctx) error {
	deviceID := c.Params("device_id")
	if err := h.DB.Where("device_id = ?", deviceID).Delete(&models.DeviceToken{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to unregister device"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}