package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func main()  {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error while loading .env")
	}
	apiKey := os.Getenv("API_KEY")
	service, err := youtube.NewService(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		panic(err)
	}
	res, err := service.Videos.
		List("snippet,statistics").
		Id("4l2jpzPDtuQ").
		Do()
	if err != nil {
		panic(err)
	}
	r, _ := res.MarshalJSON()
	fmt.Println(string(r))
}


