# Go Playground
Testing Go, Scribbling, Learning, etc

## Youtube Scraper

Scrapes all the recommended videos from the first video by DFS

YTScraper module scrapes videos recursively through the recommended videos
``` go
type Scraper interface {
	Scrape(url string, depth int, wg *sync.WaitGroup, relation int)
	getVideoInfo(doc *goquery.Document, vidInfo chan<- Video, url string)
	checkVisit(url string)
}
```
YTScraper implements the `Scraper` interface

### TODO
- ~~Goroutines~~
- Change `getVideoInfo` use the [Youtube API](https://godoc.org/google.golang.org/api/youtube/v3)
- Youtube playlist generator 

## Dependencies
- [Goquery](https://github.com/PuerkitoBio/goquery) For exploring HTML nodes