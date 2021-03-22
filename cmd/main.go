package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AleksK1NG/http-client/internal/client"
)

const (
	URL        = "https://zatey.ru/prazdnik/vozdushnye-shariki/"
	ctxTimeout = 5 * time.Second
)

func main() {
	log.Println("Starting server")
	httpClient := client.NewHttpClient()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := httpClient.CheckURL(ctx, URL)
	if err != nil {
		cancel()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case v := <-quit:
		log.Printf("signal.Notify: %v", v)
	case done := <-ctx.Done():
		log.Printf("ctx.Done: %v", done)
	}
}
