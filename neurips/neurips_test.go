package neurips_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tamnd/neurips-cli/neurips"
)

func makeHTML(titles []string, year int) string {
	var sb strings.Builder
	sb.WriteString(`<ul class="paper-list">`)
	for i, title := range titles {
		hash := fmt.Sprintf("%032d", i)
		sb.WriteString(fmt.Sprintf(
			`<li><a title="paper title" href="/paper_files/paper/%d/hash/%s-Abstract-Conference.html">%s</a></li>`,
			year, hash, title,
		))
	}
	sb.WriteString(`</ul>`)
	return sb.String()
}

func TestList(t *testing.T) {
	titles := []string{"Learning Machines", "Deep Nets", "Vision Transformers"}
	html := makeHTML(titles, 2024)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()

	cfg := neurips.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0

	c := neurips.NewClient(cfg)
	papers, err := c.List(context.Background(), 2024, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(papers) != 3 {
		t.Fatalf("got %d papers, want 3", len(papers))
	}
	if papers[0].Title != "Learning Machines" {
		t.Errorf("title = %q, want 'Learning Machines'", papers[0].Title)
	}
	if papers[0].Rank != 1 {
		t.Errorf("rank = %d, want 1", papers[0].Rank)
	}
	if papers[0].Year != 2024 {
		t.Errorf("year = %d, want 2024", papers[0].Year)
	}
}

func TestListLimit(t *testing.T) {
	titles := []string{"A", "B", "C", "D", "E"}
	html := makeHTML(titles, 2024)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()

	cfg := neurips.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0

	c := neurips.NewClient(cfg)
	papers, err := c.List(context.Background(), 2024, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(papers) != 3 {
		t.Fatalf("got %d papers, want 3 (limit applied)", len(papers))
	}
}

func TestSearch(t *testing.T) {
	titles := []string{"Machine Learning Theory", "Computer Vision", "Machine Translation"}
	html := makeHTML(titles, 2024)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()

	cfg := neurips.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0

	c := neurips.NewClient(cfg)
	papers, err := c.Search(context.Background(), "machine", 2024, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(papers) != 2 {
		t.Fatalf("got %d papers, want 2 (machine matches)", len(papers))
	}
	if papers[0].Title != "Machine Learning Theory" {
		t.Errorf("first match title = %q", papers[0].Title)
	}
	if papers[0].Rank != 1 {
		t.Errorf("rank = %d, want 1", papers[0].Rank)
	}
}

func TestYears(t *testing.T) {
	years := neurips.Years()
	if len(years) == 0 {
		t.Fatal("Years() returned empty slice")
	}
	found := false
	for _, y := range years {
		if y.Year == 2024 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Years() does not include 2024")
	}
}
