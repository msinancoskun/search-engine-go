package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"search-engine-go/internal/domain"

	"golang.org/x/time/rate"
)

type JSONProviderAdapter struct {
	name        string
	url         string
	client      *http.Client
	rateLimiter *rate.Limiter
	retryCount  int
	retryDelay  time.Duration
}

func NewJSONProviderAdapter(name, url string, rateLimit int, timeout time.Duration) *JSONProviderAdapter {
	return NewJSONProviderAdapterWithRetry(name, url, rateLimit, timeout, 3, 1*time.Second)
}

func NewJSONProviderAdapterWithRetry(name, url string, rateLimit int, timeout time.Duration, retryCount int, retryDelay time.Duration) *JSONProviderAdapter {
	rps := float64(rateLimit) / 60.0
	if rps < 1 {
		rps = 1
	}

	return &JSONProviderAdapter{
		name:        name,
		url:         url,
		client:      &http.Client{Timeout: timeout},
		rateLimiter: rate.NewLimiter(rate.Limit(rps), rateLimit),
		retryCount:  retryCount,
		retryDelay:  retryDelay,
	}
}

func (a *JSONProviderAdapter) GetName() string {
	return a.name
}

func (a *JSONProviderAdapter) GetRateLimit() int {
	return int(a.rateLimiter.Limit() * 60)
}

func (a *JSONProviderAdapter) FetchContent(ctx context.Context, query string, contentType *domain.ContentType) ([]*domain.Content, error) {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	var body []byte
	var err error

	if a.isFilePath(a.url) {
		body, err = os.ReadFile(a.url)
		if err != nil {
			return nil, fmt.Errorf("failed to read mock file: %w", err)
		}
	} else {
		reqURL := fmt.Sprintf("%s?q=%s", a.url, query)
		if contentType != nil {
			reqURL += fmt.Sprintf("&type=%s", *contentType)
		}

		var resp *http.Response
		var lastErr error

		for attempt := 0; attempt <= a.retryCount; attempt++ {
			if attempt > 0 {
				delay := a.retryDelay * time.Duration(1<<uint(attempt-1))
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(delay):
				}
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}

			req.Header.Set("Accept", "application/json")

			resp, lastErr = a.client.Do(req)
			if lastErr == nil && resp.StatusCode == http.StatusOK {
				break
			}

			if resp != nil {
				resp.Body.Close()
			}

			if resp != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return nil, fmt.Errorf("client error: status code %d", resp.StatusCode)
			}
		}

		if lastErr != nil {
			return nil, fmt.Errorf("failed to execute request after %d attempts: %w", a.retryCount+1, lastErr)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
	}

	var jsonResponse JSONProviderResponse
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	contents := make([]*domain.Content, 0, len(jsonResponse.Contents))
	for _, item := range jsonResponse.Contents {
		content := a.convertToDomain(item)
		contents = append(contents, content)
	}

	return contents, nil
}

func (a *JSONProviderAdapter) isFilePath(url string) bool {
	return !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")
}

type JSONProviderResponse struct {
	Contents   []JSONContentItem `json:"contents"`
	Pagination struct {
		Total   int `json:"total"`
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
	} `json:"pagination"`
}

type JSONContentItem struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Type        string   `json:"type"`
	Metrics     Metrics  `json:"metrics"`
	PublishedAt string   `json:"published_at"`
	Tags        []string `json:"tags"`
}

type Metrics struct {
	Views       int    `json:"views,omitempty"`
	Likes       int    `json:"likes,omitempty"`
	Duration    string `json:"duration,omitempty"`
	ReadingTime int    `json:"reading_time,omitempty"`
	Reactions   int    `json:"reactions,omitempty"`
}

func (a *JSONProviderAdapter) convertToDomain(item JSONContentItem) *domain.Content {
	contentType := domain.ContentTypeText
	if item.Type == "video" {
		contentType = domain.ContentTypeVideo
	}

	createdAt := time.Now()
	if item.PublishedAt != "" {
		if parsedTime, err := time.Parse(time.RFC3339, item.PublishedAt); err == nil {
			createdAt = parsedTime
		}
	}

	return &domain.Content{
		ProviderID:  fmt.Sprintf("%s_%s", a.name, item.ID),
		Provider:    a.name,
		Title:       item.Title,
		Type:        contentType,
		Views:       item.Metrics.Views,
		Likes:       item.Metrics.Likes,
		ReadingTime: item.Metrics.ReadingTime,
		Reactions:   item.Metrics.Reactions,
		CreatedAt:   createdAt,
	}
}
