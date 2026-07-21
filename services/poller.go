package services

import (
	"log"
	"sync"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"popsdaily/models"
)

// StartPoller schedules a recurring job (cron spec, e.g. "@every 15m")
// that fetches every active feed concurrently. Returns the cron instance
// so the caller can Stop() it on shutdown if needed.
func StartPoller(db *gorm.DB, spec string) *cron.Cron {
	c := cron.New()

	_, err := c.AddFunc(spec, func() {
		PollAllFeeds(db)
	})
	if err != nil {
		log.Fatalf("failed to schedule poller: %v", err)
	}

	c.Start()
	log.Printf("poller scheduled with spec: %s", spec)
	return c
}

// PollAllFeeds fetches every active feed concurrently, logs a summary, and
// pushes a notification for any category that got new articles this run.
func PollAllFeeds(db *gorm.DB) {
	var feeds []models.Feed
	if err := db.Where("active = ?", true).Find(&feeds).Error; err != nil {
		log.Printf("poller: failed to load feeds: %v", err)
		return
	}

	var wg sync.WaitGroup
	for i := range feeds {
		wg.Add(1)
		go func(feed models.Feed) {
			defer wg.Done()
			inserted, err := FetchAndStoreFeed(db, &feed)
			if err != nil {
				log.Printf("poller: %s failed: %v", feed.Name, err)
				return
			}
			log.Printf("poller: %s -> %d new articles", feed.Name, inserted)

			if inserted > 0 {
				NotifyNewArticles(db, feed.Category, inserted)
			}
		}(feeds[i])
	}
	wg.Wait()
}