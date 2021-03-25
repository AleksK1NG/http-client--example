package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

const (
	clientTimeout             = 3 * time.Second
	dialContextTimeout        = 5 * time.Second
	clientTLSHandshakeTimeout = 5 * time.Second
	defaultMaxIdleConnections = 5
	defaultResponseTimeout    = 5 * time.Second
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
			MaxIdleConnsPerHost:   defaultMaxIdleConnections,
			ResponseHeaderTimeout: defaultResponseTimeout,
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
	atomic.AddInt64(&h.counter, 1)
}

// GetCounter return current counter value
func (h *httpClient) GetCounter() int64 {
	return atomic.LoadInt64(&h.counter)
}

// get makes GET request for given url and writes timeoutErrChan or errChan depends on error type
func (h *httpClient) get(ctx context.Context, addressURL string) error {
	defer h.wg.Done()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addressURL, nil)
	if err != nil {
		h.errChan <- err
		return nil
	}

	res, err := h.client.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			if urlErr.Timeout() {
				h.timeoutErrChan <- urlErr.Err
				return err
			}
		}
		h.errChan <- err
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		resErr := fmt.Errorf("error status code: %v", res.StatusCode)
		h.errChan <- resErr
		return nil
	}
	return nil
}

// spamURL concurrent spam requests for given url
func (h *httpClient) spamURL(ctx context.Context, url string) {
	errGroup, ctx := errgroup.WithContext(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			h.logger.Info().Msgf("sending %v concurrent requests 💫", h.GetCounter())

			for i := 0; i < int(h.GetCounter()); i++ {
				h.wg.Add(1)
				errGroup.Go(func() error {
					return h.get(ctx, url)
				})
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
