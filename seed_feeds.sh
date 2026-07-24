#!/bin/bash

API="http://localhost:8080/api/feeds"

feeds=(


# BBC News
  "BBC Africa|https://feeds.bbci.co.uk/news/world/africa/rss.xml|national"
  "BBC World News|https://feeds.bbci.co.uk/news/world/rss.xml|world"
  "BBC Business|https://feeds.bbci.co.uk/news/business/rss.xml|business"
  "BBC Sport|https://feeds.bbci.co.uk/sport/rss.xml|sports"
  "BBC Technology|https://feeds.bbci.co.uk/news/technology/rss.xml|technology"
  "BBC Entertainment & Arts|https://feeds.bbci.co.uk/news/entertainment_and_arts/rss.xml|entertainment"


  # Daily Sun
  "Daily Sun National|https://thesun.ng/category/national/feed|national"
  "Daily Sun Politics|https://thesun.ng/category/politics/feed|politics"
  "Daily Sun Business|https://thesun.ng/category/business/feed|business"
  "Daily Sun Sports|https://thesun.ng/category/sports/feed|sports"
  "Daily Sun Entertainment|https://thesun.ng/category/entertainment/feed|entertainment"
  "Daily Sun Technology|https://thesun.ng/category/technology/feed|technology"

  # Vanguard
  "Vanguard Main|https://www.vanguardngr.com/feed/|main"
  "Vanguard National News|https://www.vanguardngr.com/category/national-news/feed/|national"
  "Vanguard Politics|https://www.vanguardngr.com/category/politics/feed/|politics"
  "Vanguard Business|https://www.vanguardngr.com/category/business/feed/|business"
  "Vanguard Sports|https://www.vanguardngr.com/category/sports/feed/|sports"
  "Vanguard Entertainment|https://www.vanguardngr.com/category/entertainment/feed/|entertainment"
  "Vanguard Metro|https://www.vanguardngr.com/category/metro/feed/|metro"
  "Vanguard Technology|https://www.vanguardngr.com/category/technology/feed/|technology"
  "Vanguard Health|https://www.vanguardngr.com/category/health/feed/|health"
  "Vanguard Editorial|https://www.vanguardngr.com/category/editorial/feed/|editorial"

  # Punch
  "Punch Latest News|https://rss.punchng.com/v1/category/latest_news|national"
  "Punch Politics|https://rss.punchng.com/v1/category/politics|politics"
  "Punch Business|https://rss.punchng.com/v1/category/business|business"
  "Punch Sports|https://rss.punchng.com/v1/category/sports|sports"
  "Punch Entertainment|https://rss.punchng.com/v1/category/entertainment|entertainment"
  "Punch Metro Plus|https://rss.punchng.com/v1/category/metro_plus|metro"

  # Daily Post
  "Daily Post Main|https://dailypost.ng/feed/|main"
  "Daily Post Politics|https://dailypost.ng/politics/feed/|politics"
  "Daily Post Metro|https://dailypost.ng/metro/feed/|metro"
  "Daily Post Sports|https://dailypost.ng/sport/feed/|sports"
  "Daily Post Entertainment|https://dailypost.ng/entertainment/feed/|entertainment"

  # The Guardian
  "The Guardian Main|https://guardian.ng/feed/|main"
  "The Guardian Nigeria News|https://guardian.ng/category/news/nigeria/feed/|national"
  "The Guardian Politics|https://guardian.ng/category/politics/feed/|politics"
  "The Guardian Business|https://guardian.ng/category/business/feed/|business"
  "The Guardian Technology|https://guardian.ng/category/technology/feed/|technology"
)

for entry in "${feeds[@]}"; do
  IFS='|' read -r name url category <<< "$entry"
  echo "Adding: $name -> $url"
  curl -s -X POST "$API" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"$name\",\"url\":\"$url\",\"category\":\"$category\"}"
  echo ""
done

echo "Done seeding feeds."