package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	clientTimeout             = 3 * time.Second
	dialContextTimeout        = 5 * time.Second
	clientTLSHandshakeTimeout = 5 * time.Second
)

type httpClient struct {
	counter        int64
	errChan        chan error
	timeoutErrChan chan error
	client         http.Client
	wg             *sync.WaitGroup
	logger         zerolog.Logger
}

// NewHttpClient constructor
func NewHttpClient(logger zerolog.Logger) *httpClient {
	client := http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: dialContextTimeout,
			}).DialContext,
			TLSHandshakeTimeout: clientTLSHandshakeTimeout,
		},
		Timeout: clientTimeout,
	}

	return &httpClient{
		client:         client,
		errChan:        make(chan error),
		timeoutErrChan: make(chan error),
		counter:        1,
		wg:             &sync.WaitGroup{},
		logger:         logger,
	}
}

// incCounter increment concurrent requests number
func (h *httpClient) incCounter() {
	h.counter++
}

// GetCounter return current counter value
func (h *httpClient) GetCounter() int64 {
	return h.counter
}

// get makes GET request for given url and writes timeoutErrChan or errChan depends on error type
func (h *httpClient) get(ctx context.Context, addressURL string) {
	defer h.wg.Done()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addressURL, nil)
	if err != nil {
		h.errChan <- err
		return
	}

	res, err := h.client.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			if urlErr.Timeout() {
				h.timeoutErrChan <- urlErr.Err
				return
			}
		}
		h.errChan <- err
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		h.errChan <- fmt.Errorf("error status code: %v", res.StatusCode)
		return
	}
}

// spamURL concurrent spam requests for given url
func (h *httpClient) spamURL(ctx context.Context, url string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			h.logger.Info().Msgf("sending %v concurrent requests 💫", h.GetCounter())

			for i := 0; i < int(h.GetCounter()); i++ {
				h.wg.Add(1)
				go h.get(ctx, url)
			}
			h.wg.Wait()
			h.incCounter()
		}
	}
}

// Run start http client
func (h *httpClient) Run(url string) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	go h.spamURL(ctx, url)

	for {
		select {
		case err := <-h.errChan:
			h.logger.Error().Caller().Stack().Msgf("received error: %v 👀", err)
			// return err 👀
		case err := <-h.timeoutErrChan:
			h.logger.Error().Caller().Stack().Msgf("received timeout error: %v 👀", err)
			return err
		}
	}
}
