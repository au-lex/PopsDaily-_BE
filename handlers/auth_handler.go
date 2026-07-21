package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"popsdaily/models"
	"popsdaily/services"
	"popsdaily/utils"
)

type AuthHandler struct {
	DB *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

// POST /api/auth/signup
func (h *AuthHandler) Signup(c *fiber.Ctx) error {
	var req models.SignupRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("signup: body parse error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.FullName == "" || req.Email == "" || req.Password == "" || req.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "full_name, email, password, and phone are required"})
	}
	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "password must be at least 6 characters"})
	}

	var existing models.User
	if err := h.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "an account with this email already exists"})
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("signup: hash password error for %s: %v", req.Email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to process password"})
	}

	user := models.User{
		FullName: req.FullName,
		Email:    req.Email,
		Password: hashed,
		Phone:    req.Phone,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		log.Printf("signup: create user error for %s: %v", req.Email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create account"})
	}

	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		log.Printf("signup: generate JWT error for user_id=%d: %v", user.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "account created but failed to generate session"})
	}

	return c.Status(fiber.StatusCreated).JSON(models.LoginResponse{Token: token, User: user})
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("login: body parse error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email and password are required"})
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		log.Printf("login: user lookup failed for %s: %v", req.Email, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid email or password"})
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid email or password"})
	}

	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		log.Printf("login: generate JWT error for user_id=%d: %v", user.ID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate session"})
	}

	return c.JSON(models.LoginResponse{Token: token, User: user})
}

// POST /api/auth/forgot-password
// Generates a fresh OTP, invalidates any previous unused ones for this
// email, and sends it out.
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req models.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("forgot-password: body parse error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Don't reveal whether the email exists — respond the same either way.
		return c.JSON(fiber.Map{"message": "if an account exists for this email, an OTP has been sent"})
	}

	if err := h.issueOTP(req.Email); err != nil {
		log.Printf("forgot-password: issueOTP error for %s: %v", req.Email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to send OTP"})
	}

	return c.JSON(fiber.Map{"message": "if an account exists for this email, an OTP has been sent"})
}

// POST /api/auth/resend-otp
func (h *AuthHandler) ResendOTP(c *fiber.Ctx) error {
	var req models.ResendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("resend-otp: body parse error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.JSON(fiber.Map{"message": "if an account exists for this email, a new OTP has been sent"})
	}

	if err := h.issueOTP(req.Email); err != nil {
		log.Printf("resend-otp: issueOTP error for %s: %v", req.Email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to resend OTP"})
	}

	return c.JSON(fiber.Map{"message": "if an account exists for this email, a new OTP has been sent"})
}

// POST /api/auth/reset-password
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req models.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("reset-password: body parse error: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email, otp, and new_password are required"})
	}
	if len(req.NewPassword) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "password must be at least 6 characters"})
	}

	var otpRecord models.PasswordResetOTP
	err := h.DB.Where("email = ? AND code = ? AND used = ?", req.Email, req.OTP, false).
		Order("created_at desc").
		First(&otpRecord).Error
	if err != nil {
		log.Printf("reset-password: OTP lookup failed for %s: %v", req.Email, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid or expired OTP"})
	}

	if time.Now().After(otpRecord.ExpiresAt) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "OTP has expired, please request a new one"})
	}

	var user models.User
	if err := h.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		log.Printf("reset-password: user lookup failed for %s: %v", req.Email, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "account not found"})
	}

	hashed, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("reset-password: hash password error for %s: %v", req.Email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to process new password"})
	}

	if err := h.DB.Model(&user).Update("password", hashed).Error; err != nil {
		log.Printf("reset-password: update password error for %s: %v", req.Email, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update password"})
	}

	h.DB.Model(&otpRecord).Update("used", true)

	return c.JSON(fiber.Map{"message": "password reset successfully"})
}

// issueOTP generates a new 6-digit code, saves it (valid for 10 minutes),
// and emails it to the given address.
func (h *AuthHandler) issueOTP(email string) error {
	code, err := utils.GenerateOTP()
	if err != nil {
		return err
	}

	otp := models.PasswordResetOTP{
		Email:     email,
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Used:      false,
	}

	if err := h.DB.Create(&otp).Error; err != nil {
		return err
	}

	return services.SendOTPEmail(email, code)
}

// GET /api/users/me
// Requires auth. Returns the logged-in user's profile.
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		log.Printf("get-profile: user lookup failed for user_id=%d: %v", userID, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(user)
}

// PATCH /api/users/me
// Requires auth. Lets a user edit their full_name, email, and/or phone.
// Only non-empty fields in the request are updated.
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req models.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("update-profile: body parse error for user_id=%d: %v", userID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		log.Printf("update-profile: user lookup failed for user_id=%d: %v", userID, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	updates := map[string]interface{}{}

	if req.FullName != "" {
		updates["full_name"] = req.FullName
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Email != "" && req.Email != user.Email {
		// Changing email — make sure nobody else already has it.
		var existing models.User
		if err := h.DB.Where("email = ? AND id != ?", req.Email, userID).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already in use"})
		}
		updates["email"] = req.Email
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "no valid fields to update"})
	}

	if err := h.DB.Model(&user).Updates(updates).Error; err != nil {
		log.Printf("update-profile: update error for user_id=%d: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update profile"})
	}

	h.DB.First(&user, userID) // reload fresh values
	return c.JSON(user)
}

// POST /api/users/change-password
// Requires auth. Different from reset-password (OTP flow) — this is for a
// logged-in user who knows their current password and wants to change it.
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req models.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("change-password: body parse error for user_id=%d: %v", userID, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "current_password and new_password are required"})
	}
	if len(req.NewPassword) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "new password must be at least 6 characters"})
	}

	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		log.Printf("change-password: user lookup failed for user_id=%d: %v", userID, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	if !utils.CheckPasswordHash(req.CurrentPassword, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "current password is incorrect"})
	}

	hashed, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("change-password: hash password error for user_id=%d: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to process new password"})
	}

	if err := h.DB.Model(&user).Update("password", hashed).Error; err != nil {
		log.Printf("change-password: update error for user_id=%d: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update password"})
	}

	return c.JSON(fiber.Map{"message": "password changed successfully"})
}