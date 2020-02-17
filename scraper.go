package scraper

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type set map[string]bool

// Video contains information about the Youtube video
type Video struct {
	url      string
	title    string
	views    int
	likes    int
	dislikes int
}

// RelationScraper scrapes Youtube links by recommendations
type RelationScraper struct {
	baseURL       string
	firstURL      string
	relatedVideos []Video
	visited       set
	vidInfo       chan Video
}

// Scraper is an interface for scraping websites
type Scraper interface {
	Scrape(url string, depth int, wg *sync.WaitGroup, relation int)
	getVideoInfo(doc *goquery.Document, vidInfo chan<- Video, url string)
	checkVisit(url string)
}

var (
	rwm = sync.RWMutex{}
)

func main() {
	YTScraper := RelationScraper{
		baseURL:       "https://www.youtube.com",
		relatedVideos: make([]Video, 0),
		visited:       make(set, 0),
		vidInfo:       make(chan Video),
	}
	defer close(YTScraper.vidInfo)
	fmt.Scanln(&YTScraper.firstURL)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go YTScraper.Scrape(YTScraper.firstURL, 3, &wg, 10)
	wg.Wait()

	for _, videos := range YTScraper.relatedVideos {
		fmt.Println(videos)
	}
}

// Scrape scrapes all the related videos by BFS
// Sends its url through channel to check it is scraped sends empty string if depth is 0
func (r *RelationScraper) Scrape(url string, depth int, wg *sync.WaitGroup, relation int) {
	rwm.RLock()
	_, ok := r.visited[url]
	rwm.RUnlock()
	if ok || depth <= 0 {
		wg.Done()
		return
	}
	rwm.Lock()
	r.visited[url] = true
	rwm.Unlock()

	res, err := http.Get(url)
	nwg := sync.WaitGroup{}
	checkRes(res)
	checkErr(err)
	checkRes(res)

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	go r.GetVideoInfo(doc, r.vidInfo, url)
	if info := <-r.vidInfo; info.title != "" {
		r.relatedVideos = append(r.relatedVideos, info)
	}

	doc.Find("div#content").
		Each(func(index int, s *goquery.Selection) {
			s.Find("a.content-link").Each(func(index int, link *goquery.Selection) {
				if index < relation {
					nextLink, _ := link.Attr("href")
					nwg.Add(1)
					go r.Scrape(r.baseURL+nextLink, depth-1, &nwg, relation)
				}
			})
		})
	nwg.Wait()
	wg.Done()
}

// GetVideoInfo scrapes informations of current video (title, views, category, likes, etc...)
func (r *RelationScraper) GetVideoInfo(doc *goquery.Document, vidInfo chan<- Video, url string) {
	info := Video{}
	doc.Find("div#content").
		Each(func(index int, vid *goquery.Selection) {
			title, _ := vid.Find("span.watch-title").Attr("title")
			views := stringToInt(vid.Find("div.watch-view-count").Text())
			likes := stringToInt(vid.Find(".like-button-renderer-like-button span").First().Text())
			dislikes := stringToInt(vid.Find(".like-button-renderer-dislike-button span").First().Text())
			info = Video{
				url:      url,
				title:    title,
				views:    views,
				likes:    likes,
				dislikes: dislikes,
			}
		})
	vidInfo <- info
}

func (r *RelationScraper) checkVisit(url string) {
	if len(url) == 0 {
		return
	}
	rwm.RLock()
	if visit, ok := r.visited[url]; !visit || !ok {
		fmt.Println(visit, ok)
		panic("Video is not scraped while visiting" + url)
	}
	rwm.RUnlock()
}

func stringToInt(s string) int {
	if len(s) == 0 {
		return 0
	}
	re := regexp.MustCompile("[0-9]")
	parsed := strings.Join(re.FindAllString(s, -1), "")
	result, err := strconv.Atoi(parsed)
	checkErr(err)
	return result
}

func checkRes(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request Failed with status", res.StatusCode)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
