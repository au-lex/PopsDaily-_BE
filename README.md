# rss-backend

Go + Fiber + GORM (Postgres) service that aggregates RSS/Atom feeds, dedupes
articles, and serves them over a JSON API. Includes a cron-based background
poller so feeds refresh automatically.

## Stack
- Fiber (HTTP)
- GORM + Postgres
- gofeed (RSS/Atom/JSON feed parsing)
- robfig/cron (scheduled polling)

## Setup

1. Create a Postgres database:
   ```
   createdb rss_backend
   ```

2. Copy the env file and adjust credentials:
   ```
   cp .env.example .env
   ```

3. Install dependencies:
   ```
   go mod tidy
   ```

4. Run it:
   ```
   go run main.go
   ```

The server starts on `:8080` (or `$PORT`), auto-migrates the `feeds` and
`articles` tables, and kicks off a poller that fetches every active feed on
the `POLL_INTERVAL` cron schedule (default every 15 minutes).

## API

### Feeds

| Method | Path | Description |
|---|---|---|
| POST | `/api/feeds` | Add a new feed source |
| GET | `/api/feeds` | List feeds (`?category=`, `?active=true`) |
| GET | `/api/feeds/:id` | Get a single feed |
| PATCH | `/api/feeds/:id` | Update name/url/category/active |
| DELETE | `/api/feeds/:id` | Remove a feed |
| POST | `/api/feeds/:id/fetch` | Fetch that one feed right now (synchronous, good for Postman testing) |
| POST | `/api/feeds/fetch-all` | Trigger a poll of all active feeds (async, fire-and-forget) |

**Create feed example**
```json
POST /api/feeds
{
  "name": "Punch Nigeria",
  "url": "https://punchng.com/feed",
  "category": "national"
}
```

### Articles

| Method | Path | Description |
|---|---|---|
| GET | `/api/articles` | Paginated article list |
| GET | `/api/articles/:id` | Single article with feed info |

**Query params for `/api/articles`:**
- `page` (default 1)
- `limit` (default 20, max 100)
- `feed_id` — filter to one feed
- `category` — filter by category
- `search` — matches title/description (case-insensitive)

### Health
`GET /health` → `{ "status": "ok" }`

## Suggested test flow in Postman

1. `POST /api/feeds` a couple of Nigerian outlets (e.g. Punch, Vanguard, Daily Post).
2. `POST /api/feeds/:id/fetch` on each one — this runs synchronously so you'll
   immediately see how many new articles were inserted, and you'll catch dead
   feed URLs right away (check the response error / the feed's `last_error` field).
3. `GET /api/articles?limit=10` to inspect the shape of what's stored — this is
   the structure your Flutter `ArticleScreen` models should match.
4. Iterate on the `Article` model fields (add/remove fields) based on what the
   UI actually needs before you wire up the real screens.

## Notes on dedupe

Articles are deduped by `GUID`. If a feed item doesn't provide one, we hash
the `link` as a fallback GUID so re-fetching a feed never creates duplicates.
