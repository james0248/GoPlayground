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

type video struct {
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
	relatedVideos []video
	visited       set
	vidInfo       chan video
}

// Scraper is an interface for scraping websites
type Scraper interface {
	Scrape(url string, depth int, wg *sync.WaitGroup, relation int)
	getVideoInfo(doc *goquery.Document, vidInfo chan<- video, url string)
	checkVisit(url string)
}

var rwm = sync.RWMutex{}

// NewRelationScraper returns a pointer of a new RelationScraper
func NewRelationScraper(baseURL string, firstURL string) *RelationScraper {
	r := RelationScraper{
		baseURL:       baseURL,
		firstURL:      firstURL,
		relatedVideos: make([]video, 0),
		visited:       make(set, 0),
		vidInfo:       make(chan video),
	}
	return &r
}

// Scrape uses the recScrape function to scrape videos
// Called for starting the recursion
func (rs *RelationScraper) Scrape() {
	defer close(rs.vidInfo)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go rs.recScrape(rs.firstURL, 3, &wg, 10)
	wg.Wait()
}

// PrintScrapedVideos prints all the videos that are scraped
func (rs *RelationScraper) PrintScrapedVideos() {
	for _, videos := range rs.relatedVideos {
		fmt.Println(videos)
	}
}

// recScrape scrapes all the related videos by DFS
// Sends its url through channel to check it is scraped sends empty string if depth is 0
func (rs *RelationScraper) recScrape(url string, depth int, wg *sync.WaitGroup, relation int) {
	defer wg.Done()
	rwm.RLock()
	_, ok := rs.visited[url]
	rwm.RUnlock()
	if ok || depth <= 0 {
		return
	}
	rwm.Lock()
	rs.visited[url] = true
	rwm.Unlock()

	res, err := http.Get(url)
	checkRes(res)
	checkErr(err)
	defer res.Body.Close()

	nwg := sync.WaitGroup{}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	go rs.getVideoInfo(doc, rs.vidInfo, url)
	if info := <-rs.vidInfo; info.title != "" {
		rs.relatedVideos = append(rs.relatedVideos, info)
	}

	doc.Find("div#content").
		Each(func(index int, s *goquery.Selection) {
			s.Find("a.content-link").Each(func(index int, link *goquery.Selection) {
				if index < relation {
					nextLink, _ := link.Attr("href")
					nwg.Add(1)
					go rs.recScrape(rs.baseURL+nextLink, depth-1, &nwg, relation)
				}
			})
		})
	nwg.Wait()
}

// GetVideoInfo scrapes informations of current video (title, views, category, likes, etc...)
func (rs *RelationScraper) getVideoInfo(doc *goquery.Document, vidInfo chan<- video, url string) {
	info := video{}
	doc.Find("div#content").
		Each(func(index int, vid *goquery.Selection) {
			title, _ := vid.Find("span.watch-title").Attr("title")
			views := stringToInt(vid.Find("div.watch-view-count").Text())
			likes := stringToInt(vid.Find(".like-button-renderer-like-button span").First().Text())
			dislikes := stringToInt(vid.Find(".like-button-renderer-dislike-button span").First().Text())
			info = video{
				url:      url,
				title:    title,
				views:    views,
				likes:    likes,
				dislikes: dislikes,
			}
		})
	vidInfo <- info
}

func (rs *RelationScraper) checkVisit(url string) {
	if len(url) == 0 {
		return
	}
	rwm.RLock()
	if visit, ok := rs.visited[url]; !visit || !ok {
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
