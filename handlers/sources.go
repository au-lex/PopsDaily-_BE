package handlers

// Fixed publisher IDs — small, closed set. Single source of truth,
// used by both FeedHandler.ListSources and ArticleHandler.ListArticlesBySourceID.
var sourceIDMap = map[int]string{
	1: "Vanguard",
	2: "Punch",
	// 3: "Premium Times",
	4: "Daily Post",
	5: "The Guardian",
	6: "Daily Sun",
}

// reverseSourceIDMap lets us look up an id given a source name.
var reverseSourceIDMap = func() map[string]int {
	m := make(map[string]int, len(sourceIDMap))
	for id, name := range sourceIDMap {
		m[name] = id
	}
	return m
}()