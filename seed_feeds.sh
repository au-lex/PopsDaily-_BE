#!/bin/bash
# This script adds Daily Sun RSS feeds to your Go backend.
# Make sure your Go backend (go run main.go) is running FIRST.

API="http://localhost:8080/api/feeds"

feeds=(
  "Daily Sun National|https://thesun.ng/category/national/feed|national"
  "Daily Sun Politics|https://thesun.ng/category/politics/feed|politics"
  "Daily Sun Business|https://thesun.ng/category/business/feed|business"
  "Daily Sun Sports|https://thesun.ng/category/sports/feed|sports"
  "Daily Sun Entertainment|https://thesun.ng/category/entertainment/feed|entertainment"
  "Daily Sun Technology|https://thesun.ng/category/technology/feed|technology"
)

for entry in "${feeds[@]}"; do
  IFS='|' read -r name url category <<< "$entry"
  echo "Adding: $name -> $url"
  curl -s -X POST "$API" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"$name\",\"url\":\"$url\",\"category\":\"$category\"}"
  echo -e "\n---"
done

echo "Done. Run this to check it worked:"
echo "curl http://localhost:8080/api/feeds"