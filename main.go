package main

import (
	"fmt"

	"github.com/james0248/goplayground/scraper"
)

func main() {
	firstURL := ""
	fmt.Scanln(&firstURL)
	YTScraper := scraper.NewRelationScraper("https://www.youtube.com", firstURL)
	YTScraper.Scrape(3, 10)
	YTScraper.PrintScrapedVideos()
}
