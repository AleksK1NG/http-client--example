package client

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

const (
	timeout = 5 * time.Second
)

type httpClient struct {
	counter int64
	errChan chan error
	client  http.Client
}

func NewHttpClient() *httpClient {
	client := http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: timeout,
			}).DialContext,
			TLSHandshakeTimeout: timeout,
		},
		Timeout: 1 * time.Second,
	}

	return &httpClient{client: client, errChan: make(chan error), counter: 1}
}

func (h *httpClient) IncCounter() {
	atomic.AddInt64(&h.counter, 1)
}

func (h *httpClient) Get(wg *sync.WaitGroup, addressURL string, requestID int) {
	defer wg.Done()
	req, err := http.NewRequest(http.MethodGet, addressURL, nil)
	if err != nil {
		h.errChan <- err
		return
	}

	res, err := h.client.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			if urlErr.Timeout() {
				log.Printf("Timeout error: %v", urlErr.Error())
				// return
			}
		}
		h.errChan <- err
		return
	}

	if res.StatusCode != 200 {
		log.Printf("response ERRPR: %v, rquestID: %v", res.StatusCode, requestID)
	}
	log.Printf("response: %v, rquestID: %v", res.StatusCode, requestID)
}

func (h *httpClient) SpamURL(url string) {
	wg := &sync.WaitGroup{}
	for i := 0; i < int(h.counter); i++ {
		wg.Add(1)
		go h.Get(wg, url, i)
	}
	wg.Wait()

	h.IncCounter()
}

func (h *httpClient) CheckURL(ctx context.Context, url string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-h.errChan:
			log.Printf("recieve err chan: %v", err)
			return err
		default:
			log.Printf("Counter: %v", h.counter)
			h.SpamURL(url)
		}
	}
}
