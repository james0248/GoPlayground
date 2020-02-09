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

var baseURL string = "https://www.youtube.com"
var visited set = make(set)
var relatedVideos = make([]video, 0)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	ScrapeVideo(scanner.Text(), 2)
	for _, vids := range relatedVideos {
		fmt.Println(vids)
	}
}

// ScrapeVideo scrapes all the related videos recursively
func ScrapeVideo(url string, depth int) {
	if _, visit := visited[url]; depth <= 0 || visit {
		return
	}
	visited[url] = true
	res, err := http.Get(url)
	checkRes(res)
	checkErr(err)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	GetVideoInfo(doc)
	doc.Find("div#content").
		Each(func(index int, s *goquery.Selection) {
			s.Find("a.content-link").Each(func(index int, link *goquery.Selection) {
				nextLink, _ := link.Attr("href")
				ScrapeVideo(baseURL+nextLink, depth-1)
			})
		})
}

// GetVideoInfo scrapes informations of current video (title, views, category, likes, etc...)
func GetVideoInfo(doc *goquery.Document) {
	doc.Find("div#content").
		Each(func(index int, vid *goquery.Selection) {
			title, _ := vid.Find("span.watch-title").Attr("title")
			category := vid.Find("ul.watch-info-tag-list a").Text()
			views := stringToInt(vid.Find("div.watch-view-count").Text())
			likes := stringToInt(vid.Find(".like-button-renderer-like-button span").First().Text())
			dislikes := stringToInt(vid.Find(".like-button-renderer-dislike-button span").First().Text())
			relatedVideos = append(relatedVideos, video{
				title:    title,
				category: category,
				views:    views,
				likes:    likes,
				dislikes: dislikes,
			})
		})
}

func stringToInt(s string) int {
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
		log.Fatalln(err)
	}
}
