package scraper

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type set map[string]bool

type video struct {
	id    string
	title string
	views uint64
	likes uint64
}

// RelationScraper scrapes Youtube links by recommendations
type RelationScraper struct {
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
func NewRelationScraper(firstURL string) *RelationScraper {
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
	go rs.recScrape(getID(rs.firstURL), depth, &wg, relation)
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
func (rs *RelationScraper) recScrape(videoID string, depth int, wg *sync.WaitGroup, relation int) {
	defer wg.Done()
	rwm.RLock()
	_, ok := rs.visited[videoID]
	rwm.RUnlock()
	if ok || depth <= 0 {
		return
	}
	rwm.Lock()
	rs.visited[videoID] = true
	rwm.Unlock()

	nwg := sync.WaitGroup{}
	go rs.getVideoInfo(rs.vidInfo, videoID)
	if info := <-rs.vidInfo; info.id != "" {
		rs.relatedVideos = append(rs.relatedVideos, info)
	} else {
		return
	}

	res, err := rs.service.Search.List("id").
		RelatedToVideoId(videoID).
		Type("video").
		MaxResults(int64(relation)).
		Do()
	checkErr(err)

	for _, relVideoId := range res.Items {
		nwg.Add(1)
		go rs.recScrape(relVideoId.Id.VideoId, depth-1, &nwg, relation)
	}
	nwg.Wait()
}

// GetVideoInfo scrapes informations of current video (title, views, category, likes, etc...)
func (rs *RelationScraper) getVideoInfo(vidInfo chan<- video, videoID string) {
	info := video{}
	res, err := rs.service.Videos.
		List("snippet,statistics").
		Id(videoID).
		Do()
	if err != nil || len(res.Items) == 0 {
		vidInfo <- info
		return
	}
	v := res.Items[0]
	info = video{
		id:    v.Id,
		title: v.Snippet.Title,
		views: v.Statistics.ViewCount,
		likes: v.Statistics.LikeCount,
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

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
