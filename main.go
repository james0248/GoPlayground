package main

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
type ytChannel struct {
	name        string
	subscribers int
	videos      []Video
}
type tuple struct {
	url   string
	depth int
}

var (
	baseURL       = "https://www.youtube.com"
	visited       = make(set)
	relatedVideos = make([]Video, 0)
	rwm           = sync.RWMutex{}
	vidInfo       = make(chan Video)
)

func main() {
	defer close(vidInfo)
	firstURL := ""
	fmt.Scanln(&firstURL)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go ScrapeRelatedVideo(firstURL, 3, &wg)
	wg.Wait()

	for _, videos := range relatedVideos {
		fmt.Println(videos)
	}
}

// ScrapeRelatedVideo scrapes all the related videos by BFS
// Sends its url through channel to check it is scraped sends empty string if depth is 0
func ScrapeRelatedVideo(url string, depth int, wg *sync.WaitGroup) {
	rwm.RLock()
	_, ok := visited[url]
	rwm.RUnlock()
	if ok || depth <= 0 {
		wg.Done()
		return
	}
	rwm.Lock()
	visited[url] = true
	rwm.Unlock()

	res, err := http.Get(url)
	nwg := sync.WaitGroup{}
	checkRes(res)
	checkErr(err)
	checkRes(res)

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	go GetVideoInfo(doc, vidInfo, url)
	if info := <-vidInfo; info.title != "" {
		relatedVideos = append(relatedVideos, info)
	}

	doc.Find("div#content").
		Each(func(index int, s *goquery.Selection) {
			s.Find("a.content-link").Each(func(index int, link *goquery.Selection) {
				nextLink, _ := link.Attr("href")
				nwg.Add(1)
				go ScrapeRelatedVideo(baseURL+nextLink, depth-1, &nwg)
			})
		})
	nwg.Wait()
	wg.Done()
}

// GetVideoInfo scrapes informations of current video (title, views, category, likes, etc...)
func GetVideoInfo(doc *goquery.Document, vidInfo chan<- Video, url string) {
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

func checkVisit(url string) {
	if len(url) == 0 {
		return
	}
	rwm.RLock()
	if visit, ok := visited[url]; !visit || !ok {
		fmt.Println(visit, ok)
		panic("Video is not scraped while visiting" + url)
	}
	rwm.RUnlock()
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
