package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type set map[string]bool
type video struct {
	title    string
	category string
	views    int
	likes    int
	dislikes int
}
type ytChannel struct {
	name        string
	subscribers int
	videos      []video
}

var (
	baseURL       = "https://www.youtube.com"
	visited       = make(set)
	relatedVideos = make([]video, 0)
	visitedMutex  = sync.RWMutex{}
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	initChan := make(chan string)
	vInfo := make(chan video)
	defer close(vInfo)
	defer close(initChan)

	scanner.Scan()
	firstURL := scanner.Text()
	go ScrapeRelatedVideo(firstURL, 1, initChan, vInfo)
	checkVisit(<-initChan)

	for _, vids := range relatedVideos {
		fmt.Println(vids)
	}
}

// ScrapeRelatedVideo scrapes all the related videos recursively
// Sends its url through channel to check it is scraped sends empty string if depth is 0
func ScrapeRelatedVideo(url string, depth int, prevVid chan string, vInfo chan video) {
	if depth <= 0 {
		prevVid <- ""
		return
	}

	visitedMutex.RLock()
	_, ok := visited[url]
	visitedMutex.RUnlock()
	if ok {
		prevVid <- url
		return
	}

	visitedMutex.Lock()
	visited[url] = true
	visitedMutex.Unlock()

	nextVid := make(chan string)
	res, err := http.Get(url)
	checkRes(res)
	checkErr(err)
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	go GetVideoInfo(doc, vInfo)
	doc.Find("div#content").
		Each(func(index int, s *goquery.Selection) {
			s.Find("a.content-link").Each(func(index int, link *goquery.Selection) {
				nextLink, _ := link.Attr("href")
				go ScrapeRelatedVideo(baseURL+nextLink, depth-1, nextVid, vInfo)
			})
		})

	for nextLink := range nextVid {
		checkVisit(nextLink)
	}
	for videos := range vInfo {
		relatedVideos = append(relatedVideos, videos)
	}
	prevVid <- url
}

// GetVideoInfo scrapes informations of current video (title, views, category, likes, etc...)
func GetVideoInfo(doc *goquery.Document, vInfo chan<- video) {
	doc.Find("div#content").
		Each(func(index int, vid *goquery.Selection) {
			title, _ := vid.Find("span.watch-title").Attr("title")
			category := vid.Find("ul.watch-info-tag-list a").Text()
			views := stringToInt(vid.Find("div.watch-view-count").Text())
			likes := stringToInt(vid.Find(".like-button-renderer-like-button span").First().Text())
			dislikes := stringToInt(vid.Find(".like-button-renderer-dislike-button span").First().Text())
			vInfo <- video{
				title:    title,
				category: category,
				views:    views,
				likes:    likes,
				dislikes: dislikes,
			}
		})
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
	visitedMutex.RLock()
	if visit, ok := visited[url]; !visit || !ok {
		fmt.Println(visit, ok)
		log.Fatalln("Video is not scraped while visiting", url)
	}
	visitedMutex.RUnlock()
}

func checkRes(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request Failed with status", res.StatusCode)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
