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

// Video contains information about the Youtube video
type Video struct {
	title    string
	category string
	views    int
	likes    int
	dislikes int
}
type ytChannel struct {
	name        string
	subscribers int
	videos      []Video
}

var (
	baseURL       = "https://www.youtube.com"
	visited       = make(set)
	relatedVideos = make([]Video, 0)
	visitedMutex  = sync.RWMutex{}
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	firstURL := scanner.Text()
	ScrapeRelatedVideo(firstURL, 2)

	for _, vids := range relatedVideos {
		fmt.Println(vids)
	}
}

// ScrapeRelatedVideo scrapes all the related videos recursively
// Sends its url through channel to check it is scraped sends empty string if depth is 0
func ScrapeRelatedVideo(url string, depth int) {
	_, ok := visited[url]
	if ok || depth <= 0 {
		return
	}
	visited[url] = true

	res, err := http.Get(url)
	checkRes(res)
	checkErr(err)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	relatedVideos = append(relatedVideos, GetVideoInfo(doc))
	doc.Find("div#content").
		Each(func(index int, s *goquery.Selection) {
			s.Find("a.content-link").Each(func(index int, link *goquery.Selection) {
				nextLink, _ := link.Attr("href")
				ScrapeRelatedVideo(baseURL+nextLink, depth-1)
			})
		})

}

// GetVideoInfo scrapes informations of current video (title, views, category, likes, etc...)
func GetVideoInfo(doc *goquery.Document) Video {
	info := Video{}
	doc.Find("div#content").
		Each(func(index int, vid *goquery.Selection) {
			title, _ := vid.Find("span.watch-title").Attr("title")
			category := vid.Find("ul.watch-info-tag-list a").Text()
			views := stringToInt(vid.Find("div.watch-view-count").Text())
			likes := stringToInt(vid.Find(".like-button-renderer-like-button span").First().Text())
			dislikes := stringToInt(vid.Find(".like-button-renderer-dislike-button span").First().Text())
			info = Video{
				title:    title,
				category: category,
				views:    views,
				likes:    likes,
				dislikes: dislikes,
			}
		})
	return info
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
