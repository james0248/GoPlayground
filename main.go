package main

import (
	"errors"
	"fmt"
	"net/http"
)

var errorRequestFailed = errors.New("Request Failed")

type response struct {
	url    string
	status string
}

func main() {
	urls := []string{
		"https://www.airbnb.com/",
		"https://www.google.com/",
		"https://www.amazon.com/",
		"https://www.google.com/",
		"https://soundcloud.com/",
		"https://www.facebook.com/",
		"https://www.instagram.com/",
		"https://academy.nomadcoders.co/",
	}
	results := make(map[string]string)
	channel := make(chan response)
	for _, url := range urls {
		go hitURL(url, channel)
	}
	var message response
	for i := 0; i < len(urls); i++ {
		message = <-channel
		results[message.url] = message.status
	}
	for url, result := range results {
		fmt.Println(url, "Result :", result)
	}
}

func hitURL(url string, channel chan<- response) {
	resp, err := http.Get(url)
	status := "OK"
	if err != nil || resp.StatusCode >= 400 {
		status = "FAILED"
	}
	channel <- response{
		url:    url,
		status: status,
	}
}
