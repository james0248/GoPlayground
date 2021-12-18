package main

import (
	"fmt"

	"github.com/james0248/goplayground/scraper"
)

func main() {
	fmt.Print("Input starting URL: ")
	firstURL := ""
	_, err := fmt.Scanln(&firstURL)
	if err != nil {
		panic("Input failed!")
	}
	YTScraper := scraper.NewRelationScraper(firstURL)
	YTScraper.Scrape(3, 5)
	YTScraper.PrintScrapedVideos()
}
