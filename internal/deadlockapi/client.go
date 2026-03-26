package deadlockapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL   = "https://api.deadlock-api.com"
	assetsURL = "https://assets.deadlock-api.com"
)

// Client is an HTTP client for deadlock-api.com with rate limiting.
type Client struct {
	http      *http.Client
	rateLimit chan struct{}
	mu        sync.Mutex
}

// NewClient creates a new deadlock-api.com client with the given rate limit.
func NewClient(requestsPerSecond int) *Client {
	if requestsPerSecond <= 0 {
		requestsPerSecond = 10
	}

	c := &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimit: make(chan struct{}, requestsPerSecond),
	}

	// Fill the rate limiter bucket
	for range requestsPerSecond {
		c.rateLimit <- struct{}{}
	}

	// Refill tokens at a steady rate
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(requestsPerSecond))
		defer ticker.Stop()
		for range ticker.C {
			select {
			case c.rateLimit <- struct{}{}:
			default:
			}
		}
	}()

	return c
}

// get performs a rate-limited GET request with retry on 429.
func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	const maxRetries = 5

	for attempt := range maxRetries {
		// Acquire rate limit token
		select {
		case <-c.rateLimit:
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Accept", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("executing request to %s: %w", url, err)
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			backoff := time.Duration(2<<attempt) * time.Second
			select {
			case <-time.After(backoff):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API returned %d for %s: %s", resp.StatusCode, url, string(body))
		}

		return body, nil
	}

	return nil, fmt.Errorf("API rate limited after %d retries for %s", maxRetries, url)
}

// getJSON performs a rate-limited GET and unmarshals the JSON response.
func (c *Client) getJSON(ctx context.Context, url string, v any) error {
	body, err := c.get(ctx, url)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("unmarshalling response from %s: %w", url, err)
	}
	return nil
}
