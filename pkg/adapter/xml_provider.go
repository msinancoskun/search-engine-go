package adapter

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"search-engine-go/internal/domain"

	"golang.org/x/time/rate"
)

type XMLProviderAdapter struct {
	name        string
	url         string
	client      *http.Client
	rateLimiter *rate.Limiter
	retryCount  int
	retryDelay  time.Duration
}

func NewXMLProviderAdapter(name, url string, rateLimit int, timeout time.Duration) *XMLProviderAdapter {
	return NewXMLProviderAdapterWithRetry(name, url, rateLimit, timeout, 3, 1*time.Second)
}

func NewXMLProviderAdapterWithRetry(name, url string, rateLimit int, timeout time.Duration, retryCount int, retryDelay time.Duration) *XMLProviderAdapter {
	rps := float64(rateLimit) / 60.0
	if rps < 1 {
		rps = 1
	}

	return &XMLProviderAdapter{
		name:        name,
		url:         url,
		client:      &http.Client{Timeout: timeout},
		rateLimiter: rate.NewLimiter(rate.Limit(rps), rateLimit),
		retryCount:  retryCount,
		retryDelay:  retryDelay,
	}
}

func (a *XMLProviderAdapter) GetName() string {
	return a.name
}

func (a *XMLProviderAdapter) GetRateLimit() int {
	return int(a.rateLimiter.Limit() * 60)
}

func (a *XMLProviderAdapter) FetchContent(ctx context.Context, query string, contentType *domain.ContentType) ([]*domain.Content, error) {
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

			req.Header.Set("Accept", "application/xml")

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

	var xmlResponse XMLProviderResponse
	if err := xml.Unmarshal(body, &xmlResponse); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	contents := make([]*domain.Content, 0, len(xmlResponse.Items))
	for _, item := range xmlResponse.Items {
		content := a.convertToDomain(item)
		contents = append(contents, content)
	}

	return contents, nil
}

func (a *XMLProviderAdapter) isFilePath(url string) bool {
	return !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")
}

type XMLProviderResponse struct {
	XMLName xml.Name         `xml:"feed"`
	Items   []XMLContentItem `xml:"items>item"`
	Meta    struct {
		TotalCount   int `xml:"total_count"`
		CurrentPage  int `xml:"current_page"`
		ItemsPerPage int `xml:"items_per_page"`
	} `xml:"meta"`
}

type XMLContentItem struct {
	XMLName         xml.Name `xml:"item"`
	ID              string   `xml:"id"`
	Headline        string   `xml:"headline"`
	Type            string   `xml:"type"`
	Stats           XMLStats `xml:"stats"`
	PublicationDate string   `xml:"publication_date"`
	Categories      struct {
		Category []string `xml:"category"`
	} `xml:"categories"`
}

type XMLStats struct {
	Views       int    `xml:"views,omitempty"`
	Likes       int    `xml:"likes,omitempty"`
	Duration    string `xml:"duration,omitempty"`
	ReadingTime int    `xml:"reading_time,omitempty"`
	Reactions   int    `xml:"reactions,omitempty"`
	Comments    int    `xml:"comments,omitempty"`
}

func (a *XMLProviderAdapter) convertToDomain(item XMLContentItem) *domain.Content {
	contentType := domain.ContentTypeText
	if item.Type == "video" {
		contentType = domain.ContentTypeVideo
	}

	createdAt := time.Now()
	if item.PublicationDate != "" {
		if parsedTime, err := time.Parse("2006-01-02", item.PublicationDate); err == nil {
			createdAt = parsedTime
		}
	}

	return &domain.Content{
		ProviderID:  fmt.Sprintf("%s_%s", a.name, item.ID),
		Provider:    a.name,
		Title:       item.Headline,
		Type:        contentType,
		Views:       item.Stats.Views,
		Likes:       item.Stats.Likes,
		ReadingTime: item.Stats.ReadingTime,
		Reactions:   item.Stats.Reactions,
		CreatedAt:   createdAt,
	}
}
