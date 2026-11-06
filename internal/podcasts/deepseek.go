package podcasts

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"
)

func (a *App) deepSeekSearch(ctx context.Context, p providers, e *Episode, since, cutoff time.Time) ([]Source, error) {
	// Use the standalone search API: the number of paid searches is bounded here,
	// independently of whether the writing model supports server-side tools.
	queries := []string{"technology news " + clipUTF8(e.Settings.Topics, 45) + " after:" + since.Format("2006-01-02"), "big tech AI site:reuters.com OR site:apnews.com after:" + since.Format("2006-01-02")}
	for _, id := range e.Settings.People {
		for _, who := range People {
			if who.ID == id {
				queries = append(queries, "site:x.com/"+who.Handle+" after:"+since.Format("2006-01-02"))
			}
		}
	}
	out := []Source{}
	for _, query := range queries {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		b, err := a.postTo(ctx, "https://api.deepseek.com", p.DeepSeekKey, "/v1/tools/search", map[string]any{"query": query, "count": 5, "need_content": true}, 5<<20, nil, false)
		if err != nil {
			return nil, err
		}
		ss, err := parseDeepSeekSearch(b, since, cutoff)
		if err != nil {
			return nil, err
		}
		out = append(out, ss...)
	}
	return out, nil
}
func parseDeepSeekSearch(b []byte, since, cutoff time.Time) ([]Source, error) {
	var result struct {
		Results []struct {
			Title    string `json:"title"`
			URL      string `json:"url"`
			Summary  string `json:"summary"`
			Content  string `json:"content"`
			Snippet  string `json:"snippet"`
			SiteName string `json:"site_name"`
			Date     string `json:"date"`
		} `json:"results"`
	}
	if json.Unmarshal(b, &result) != nil || result.Results == nil {
		return nil, errors.New("DeepSeek search returned unreadable evidence")
	}
	out := []Source{}
	for _, r := range result.Results {
		if !webURL(r.URL) {
			continue
		}
		when := sourceDate(r.Date)
		if !when.IsZero() && (when.Before(since) || when.After(cutoff.Add(5*time.Minute))) {
			continue
		}
		publisher := r.SiteName
		u, _ := url.Parse(r.URL)
		if publisher == "" {
			publisher = u.Hostname()
		}
		kind := "web"
		if u.Hostname() == "x.com" || u.Hostname() == "www.x.com" || u.Hostname() == "twitter.com" {
			kind = "social"
		}
		excerpt := r.Summary
		if strings.TrimSpace(excerpt) == "" {
			excerpt = r.Content
		}
		if strings.TrimSpace(excerpt) == "" {
			excerpt = r.Snippet
		}
		out = append(out, Source{Title: plain(r.Title), URL: r.URL, Publisher: publisher, Kind: kind, Published: r.Date, Excerpt: plain(excerpt)})
	}
	return out, nil
}
