package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
	"gorm.io/gorm"

	"popsdaily/models"
)

var fcmClient *messaging.Client

// InitFirebase sets up the Firebase Admin SDK. FIREBASE_SERVICE_ACCOUNT can
// be either:
//   - the raw JSON content of the service account key (what you'll use on
//     Railway, pasted as one env var value), or
//   - a file path to the downloaded .json key (handy for local dev, e.g.
//     FIREBASE_SERVICE_ACCOUNT=./firebase-service-account.json)
//
// Call this once from main.go on startup.
func InitFirebase() error {
	raw := os.Getenv("FIREBASE_SERVICE_ACCOUNT")
	if raw == "" {
		log.Println("notifications: FIREBASE_SERVICE_ACCOUNT not set, push notifications disabled")
		return nil
	}

	// If it doesn't look like JSON, treat it as a file path and read it.
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "{") {
		data, err := os.ReadFile(trimmed)
		if err != nil {
			return fmt.Errorf("FIREBASE_SERVICE_ACCOUNT looks like a file path but could not be read: %w", err)
		}
		raw = string(data)
	}

	if !json.Valid([]byte(raw)) {
		return fmt.Errorf("FIREBASE_SERVICE_ACCOUNT is not valid JSON")
	}

	opt := option.WithCredentialsJSON([]byte(raw))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return fmt.Errorf("failed to init firebase app: %w", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return fmt.Errorf("failed to init firebase messaging client: %w", err)
	}

	fcmClient = client
	log.Println("notifications: firebase initialized")
	return nil
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// NotifyNewArticles pushes a notification to every registered device telling
// them how many new articles landed for a given category. Silently no-ops
// if Firebase isn't configured, so local dev without FCM creds doesn't crash.
func NotifyNewArticles(db *gorm.DB, category string, count int) {
	if fcmClient == nil || count == 0 {
		return
	}

	var tokens []string
	if err := db.Model(&models.DeviceToken{}).Pluck("token", &tokens).Error; err != nil {
		log.Printf("notifications: failed to load device tokens: %v", err)
		return
	}
	if len(tokens) == 0 {
		return
	}

	title := "New stories in " + capitalize(category)
	body := fmt.Sprintf("%d new article(s) just dropped — tap to read", count)

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: map[string]string{
			"category": category,
		},
	}

	resp, err := fcmClient.SendEachForMulticast(context.Background(), message)
	if err != nil {
		log.Printf("notifications: failed to send push: %v", err)
		return
	}

	log.Printf("notifications: sent %d, failed %d for category %s", resp.SuccessCount, resp.FailureCount, category)

	// Clean up dead tokens (uninstalled app, expired, etc.) so the list
	// doesn't grow stale over time.
	if resp.FailureCount > 0 {
		for i, r := range resp.Responses {
			if !r.Success {
				db.Where("token = ?", tokens[i]).Delete(&models.DeviceToken{})
			}
		}
	}
}