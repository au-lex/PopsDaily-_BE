#!/bin/bash
# This script adds Daily Sun RSS feeds to your Go backend.
# Make sure your Go backend (go run main.go) is running FIRST.

# API="http://localhost:8080/api/feeds"

# feeds=(
#   "Daily Sun National|https://thesun.ng/category/national/feed|national"
#   "Daily Sun Politics|https://thesun.ng/category/politics/feed|politics"
#   "Daily Sun Business|https://thesun.ng/category/business/feed|business"
#   "Daily Sun Sports|https://thesun.ng/category/sports/feed|sports"
#   "Daily Sun Entertainment|https://thesun.ng/category/entertainment/feed|entertainment"
#   "Daily Sun Technology|https://thesun.ng/category/technology/feed|technology"
# )

# for entry in "${feeds[@]}"; do
#   IFS='|' read -r name url category <<< "$entry"
#   echo "Adding: $name -> $url"
#   curl -s -X POST "$API" \
#     -H "Content-Type: application/json" \
#     -d "{\"name\":\"$name\",\"url\":\"$url\",\"category\":\"$category\"}"
#   echo -e "\n---"
# done

# echo "Done. Run this to check it worked:"
# echo "curl http://localhost:8080/api/feeds"


#!/bin/bash
# Seeds your Go RSS backend with feeds from Vanguard, Punch, Premium Times,
# Daily Post, The Guardian, and Daily Sun.
#
# Make sure your Go backend (go run main.go) is running FIRST.
# Usage: chmod +x seed_feeds.sh && ./seed_feeds.sh

API="http://localhost:8080/api/feeds"

# Each entry: Name|URL|Category
feeds=(
  # --- Vanguard ---
  "Vanguard Main|https://www.vanguardngr.com/feed/|national"
  "Vanguard National News|https://www.vanguardngr.com/category/national-news/feed/|national"
  "Vanguard Politics|https://www.vanguardngr.com/category/politics/feed/|politics"
  "Vanguard Business|https://www.vanguardngr.com/category/business/feed/|business"
  "Vanguard Sports|https://www.vanguardngr.com/category/sports/feed/|sports"
  "Vanguard Entertainment|https://www.vanguardngr.com/category/entertainment/feed/|entertainment"
  "Vanguard Metro|https://www.vanguardngr.com/category/metro/feed/|national"
  "Vanguard Technology|https://www.vanguardngr.com/category/technology/feed/|technology"
  "Vanguard Health|https://www.vanguardngr.com/category/health/feed/|national"
  "Vanguard Editorial|https://www.vanguardngr.com/category/editorial/feed/|national"

  # --- Punch ---
  "Punch Latest News|https://rss.punchng.com/v1/category/latest_news|national"
  "Punch Politics|https://rss.punchng.com/v1/category/politics|politics"
  "Punch Business|https://rss.punchng.com/v1/category/business|business"
  "Punch Sports|https://rss.punchng.com/v1/category/sports|sports"
  "Punch Entertainment|https://rss.punchng.com/v1/category/entertainment|entertainment"
  "Punch Metro Plus|https://rss.punchng.com/v1/category/metro_plus|national"

  # --- Premium Times ---
  "Premium Times Main|https://www.premiumtimesng.com/feed/|national"
  "Premium Times News|https://www.premiumtimesng.com/category/news/feed/|national"
  "Premium Times Politics|https://www.premiumtimesng.com/category/news/politics/feed/|politics"
  "Premium Times Business|https://www.premiumtimesng.com/category/business/feed/|business"
  "Premium Times Investigations|https://www.premiumtimesng.com/category/investigations/feed/|national"

  # --- Daily Post ---
  "Daily Post Main|https://dailypost.ng/feed/|national"
  "Daily Post Politics|https://dailypost.ng/politics/feed/|politics"
  "Daily Post Metro|https://dailypost.ng/metro/feed/|national"
  "Daily Post Sports|https://dailypost.ng/sport/feed/|sports"
  "Daily Post Entertainment|https://dailypost.ng/entertainment/feed/|entertainment"

  # --- The Guardian (NG) ---
  "The Guardian Main|https://guardian.ng/feed/|national"
  "The Guardian Nigeria News|https://guardian.ng/category/news/nigeria/feed/|national"
  "The Guardian Politics|https://guardian.ng/category/politics/feed/|politics"
  "The Guardian Business|https://guardian.ng/category/business/feed/|business"
  "The Guardian Technology|https://guardian.ng/category/technology/feed/|technology"


    "Daily Sun National|https://thesun.ng/category/national/feed|national"
  "Daily Sun Politics|https://thesun.ng/category/politics/feed|politics"
  "Daily Sun Business|https://thesun.ng/category/business/feed|business"
  "Daily Sun Sports|https://thesun.ng/category/sports/feed|sports"
  "Daily Sun Entertainment|https://thesun.ng/category/entertainment/feed|entertainment"
  "Daily Sun Technology|https://thesun.ng/category/technology/feed|technology"

)

success=0
failed=0

for entry in "${feeds[@]}"; do
  IFS='|' read -r name url category <<< "$entry"
  echo "Adding: $name -> $url"

  response=$(curl -s -w "\n%{http_code}" -X POST "$API" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"$name\",\"url\":\"$url\",\"category\":\"$category\"}")

  http_code=$(echo "$response" | tail -n1)
  body=$(echo "$response" | sed '$d')

  if [[ "$http_code" =~ ^2 ]]; then
    echo "  ✓ OK ($http_code)"
    ((success++))
  else
    echo "  ✗ FAILED ($http_code): $body"
    ((failed++))
  fi
  echo "---"
done

echo ""
echo "Done. $success added, $failed failed."
echo "Verify with: curl $API"