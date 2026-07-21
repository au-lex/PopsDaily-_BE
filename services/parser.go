package services

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
	"gorm.io/gorm"

	"rss-backend/models"
)

var fp = newFeedParser()

// newFeedParser configures gofeed with a custom HTTP client. Some Nigerian
// news sites (Vanguard, etc.) sit behind a WAF/CDN that 403s requests
// carrying Go's default user-agent, so we spoof a normal browser UA and set
// a reasonable timeout.
func newFeedParser() *gofeed.Parser {
	p := gofeed.NewParser()
	p.Client = &http.Client{Timeout: 15 * time.Second}
	p.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	return p
}

// FetchAndStoreFeed pulls a single feed's URL, parses it, and upserts
// articles into the database, deduped by GUID (falling back to a hash
// of the link if the source feed doesn't provide a stable GUID).
func FetchAndStoreFeed(db *gorm.DB, feed *models.Feed) (int, error) {
	parsed, err := fp.ParseURL(feed.URL)
	now := time.Now()

	if err != nil {
		db.Model(feed).Updates(map[string]interface{}{
			"last_fetched_at": now,
			"last_error":      err.Error(),
		})
		return 0, err
	}

	inserted := 0
	for _, item := range parsed.Items {
		guid := item.GUID
		if guid == "" {
			guid = hashString(item.Link)
		}

		var existing models.Article
		result := db.Where("guid = ?", guid).First(&existing)
		if result.Error == nil {
			continue // already have it
		}

		published := now
		if item.PublishedParsed != nil {
			published = *item.PublishedParsed
		}

		article := models.Article{
			FeedID:      feed.ID,
			Title:       item.Title,
			Link:        item.Link,
			GUID:        guid,
			Description: item.Description,
			PublishedAt: published,
			Category:    feed.Category,
		}

		if item.Author != nil {
			article.Author = item.Author.Name
		}
		if item.Image != nil {
			article.ImageURL = item.Image.URL
		} else if len(item.Enclosures) > 0 {
			article.ImageURL = item.Enclosures[0].URL
		}
		if item.Content != "" {
			article.Content = item.Content
		}

		if err := db.Create(&article).Error; err != nil {
			log.Printf("failed to store article %q: %v", item.Title, err)
			continue
		}
		inserted++
	}

	db.Model(feed).Updates(map[string]interface{}{
		"last_fetched_at": now,
		"last_error":      "",
	})

	return inserted, nil
}

func hashString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}