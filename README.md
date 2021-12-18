# Go Playground

Testing Go, Scribbling, Learning, etc

## Youtube Scraper

Scrapes all the recommended videos from the first video by DFS

YTScraper module scrapes videos recursively through the recommended videos

```go
type Scraper interface {
	Scrape(videoId string, depth int, wg *sync.WaitGroup, relation int)
	getVideoInfo(vidInfo chan<- Video, url string)
}
```

YTScraper implements the `Scraper` interface

### TODO

-   ~~Goroutines~~
-   ~~Change `getVideoInfo` to use the [Youtube API](https://godoc.org/google.golang.org/api/youtube/v3)~~
-   Youtube playlist generator
-   Add option to use plain web scraping instead of Youtube API (API cost for getting related videos is very high)

## Dependencies

-   [Goquery](https://github.com/PuerkitoBio/goquery) For exploring HTML nodes

-   [Youtube API](https://godoc.org/google.golang.org/api/youtube/v3)
