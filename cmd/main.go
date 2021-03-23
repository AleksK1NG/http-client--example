package main

import (
	"flag"

	"github.com/AleksK1NG/http-client/internal/client"
	"github.com/AleksK1NG/http-client/pkg/logger"
)

var url string

func init() {
	flag.StringVar(&url, "url", "https://www.youtube.com/", "url address")
}

func main() {
	flag.Parse()
	l := logger.NewLogger()

	l.Info().Msg("Starting http client")
	httpClient := client.NewHttpClient(l)

	err := httpClient.Run(url)
	if err != nil {
		l.Info().Msgf("httpClient concurrent requests number: %v 🤓", httpClient.GetCounter())
	}
}
