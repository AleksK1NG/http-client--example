package main

import (
	"context"
	"log"
	"os"

	"github.com/AleksK1NG/http-client/internal/client"
)

const (
	URL = "https://zatey.ru/prazdnik/vozdushnye-shariki/"
)

func main() {
	log.Println("Starting server")
	httpClient := client.NewHttpClient()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := httpClient.CheckURL(ctx, URL)
	if err != nil {
		log.Printf("httpClient couter: %v, error: %v", httpClient.Counter(), err)
		os.Exit(0)
	}
}
