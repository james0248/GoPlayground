package scraper

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type set map[string]bool

type video struct {
	id       string
	title    string
	views    uint64
	likes    uint64
	dislikes uint64
}

// RelationScraper scrapes Youtube links by recommendations
type RelationScraper struct {
	baseURL       string
	firstURL      string
	relatedVideos []video
	visited       set
	vidInfo       chan video
	service       *youtube.Service
}

var (
	rwm = sync.RWMutex{}
)

// NewRelationScraper returns a pointer of a new RelationScraper
func NewRelationScraper(baseURL, firstURL string) *RelationScraper {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error while loading .env")
	}
	apiKey := os.Getenv("API_KEY")
	service, err := youtube.NewService(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalln(err, "Error while loading youtube API service")
	}
	rs := RelationScraper{
		baseURL:       baseURL,
		firstURL:      firstURL,
		relatedVideos: make([]video, 0),
		visited:       make(set, 0),
		vidInfo:       make(chan video),
		service:       service,
	}
	return &rs
}

// Scrape uses the recScrape function to scrape videos
// Called for starting the recursion
func (rs *RelationScraper) Scrape(depth, relation int) {
	defer close(rs.vidInfo)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go rs.recScrape(rs.firstURL, depth, &wg, relation)
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
	go rs.getVideoInfo(rs.vidInfo, url)
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
func (rs *RelationScraper) getVideoInfo(vidInfo chan<- video, url string) {
	info := video{}
	res, err := rs.service.Videos.
		List("snippet,statistics").
		Id(getID(url)).
		Do()
	if err != nil {
		panic(err)
	}
	v := res.Items[0]
	info = video{
		id:       v.Id,
		title:    v.Snippet.Title,
		views:    v.Statistics.ViewCount,
		likes:    v.Statistics.LikeCount,
		dislikes: v.Statistics.DislikeCount,
	}
	vidInfo <- info
}

func getID(s string) string {
	u, err := url.Parse(s)
	checkErr(err)
	m, err := url.ParseQuery(u.RawQuery)
	checkErr(err)
	return m["v"][0]
}

func stringToInt(s string) int {
	if s == "" {
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
