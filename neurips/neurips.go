// Package neurips is the library behind the neurips command line:
// the HTTP client, request shaping, and the typed data models for NeurIPS
// conference papers fetched from papers.nips.cc.
package neurips

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// AvailableYears lists NeurIPS proceedings years (newest first).
var AvailableYears = []int{2025, 2024, 2023, 2022, 2021, 2020, 2019, 2018, 2017, 2016, 2015, 2014, 2013, 2012, 2011}

const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

var paperRe = regexp.MustCompile(`title="paper title" href="(/paper_files/paper/\d+/hash/[^"]+)">(.*?)(?:</a>|\.\.\.)`)

// Config holds constructor parameters.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://papers.nips.cc",
		UserAgent: DefaultUserAgent,
		Rate:      500 * time.Millisecond,
		Retries:   3,
		Timeout:   30 * time.Second,
	}
}

// Client fetches NeurIPS paper listings.
type Client struct {
	cfg        Config
	httpClient *http.Client
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client with the given config.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		b, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return b, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get: %w", lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// List returns all papers for the given year.
func (c *Client) List(ctx context.Context, year, limit int) ([]Paper, error) {
	rawURL := fmt.Sprintf("%s/paper_files/paper/%d", c.cfg.BaseURL, year)
	raw, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	return parsePapers(string(raw), year, limit, c.cfg.BaseURL), nil
}

// Search returns papers for the given year whose title contains query (case-insensitive).
func (c *Client) Search(ctx context.Context, query string, year, limit int) ([]Paper, error) {
	all, err := c.List(ctx, year, 0)
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	var out []Paper
	rank := 0
	for _, p := range all {
		if strings.Contains(strings.ToLower(p.Title), q) {
			rank++
			p.Rank = rank
			out = append(out, p)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

// Years returns the list of available NeurIPS years as Year records.
func Years() []Year {
	out := make([]Year, len(AvailableYears))
	for i, y := range AvailableYears {
		out[i] = Year{Year: y}
	}
	return out
}

func parsePapers(html string, year, limit int, baseURL string) []Paper {
	matches := paperRe.FindAllStringSubmatch(html, -1)
	out := make([]Paper, 0, len(matches))
	for i, m := range matches {
		if limit > 0 && i >= limit {
			break
		}
		path := m[1]
		title := strings.TrimSpace(m[2])
		out = append(out, Paper{
			Rank:  i + 1,
			Year:  year,
			Title: title,
			URL:   baseURL + path,
		})
	}
	return out
}
