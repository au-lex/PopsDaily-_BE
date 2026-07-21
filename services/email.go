package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const resendAPIURL = "https://api.resend.com/emails"

type resendEmailPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

// SendOTPEmail sends the password reset code to the user's email via Resend.

func SendOTPEmail(toEmail, otp string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	from := os.Getenv("RESEND_FROM")

	if apiKey == "" || from == "" {
		return fmt.Errorf("RESEND_API_KEY or RESEND_FROM env var not set")
	}

	html := fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 480px; margin: 0 auto;">
			<h2>Password Reset Code</h2>
			<p>Your OTP code is:</p>
			<p style="font-size: 28px; font-weight: bold; letter-spacing: 4px;">%s</p>
			<p>This code expires in 10 minutes. If you didn't request this, you can ignore this email.</p>
		</div>
	`, otp)

	payload := resendEmailPayload{
		From:    from,
		To:      []string{toEmail},
		Subject: "Your password reset code",
		HTML:    html,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, resendAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("resend API returned status %d", resp.StatusCode)
	}

	return nil
}