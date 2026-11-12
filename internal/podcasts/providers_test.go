package podcasts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
)

func TestProviderSecretsMigrationAndEndpointBinding(t *testing.T) {
	a, h, c, session := fixture(t)
	payload := map[string]any{"textProvider": "deepseek", "deepSeekKey": "synthetic-deepseek-secret", "serviceURL": "http://127.0.0.1:19552", "serviceToken": strings.Repeat("s", 40)}
	if w := request(h, "PUT", "/api/providers", payload, c, session.CSRF); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	state := request(h, "GET", "/api/state", nil, c, "")
	if strings.Contains(state.Body.String(), "synthetic-deepseek-secret") || strings.Contains(state.Body.String(), strings.Repeat("s", 40)) {
		t.Fatal("returned provider secrets")
	}
	if w := request(h, "PUT", "/api/providers", map[string]string{"serviceURL": "https://example.com"}, c, session.CSRF); w.Code != 400 {
		t.Fatal("non-loopback accepted")
	}
	if w := request(h, "PUT", "/api/providers", map[string]string{"serviceURL": "http://127.0.0.1:19553"}, c, session.CSRF); w.Code != 200 || a.provider().ServiceToken != "" {
		t.Fatal("credential followed a changed destination")
	}
	// Save the old provider JSON shape, then verify a normal application restart.
	legacy := []byte(`{"key":"legacy-openai-key","xToken":"legacy-x-token","model":"gpt-5.4","ttsModel":"gpt-4o-mini-tts"}`)
	encrypted, err := a.crypt(legacy, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = h.DB.Exec("UPDATE podcast_providers SET secret=? WHERE id=1", encrypted); err != nil {
		t.Fatal(err)
	}
	_ = h.Close()
	host, err := apphost.New(h.Config, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer host.Close()
	b, err := New(host)
	if err != nil {
		t.Fatal(err)
	}
	p := b.provider()
	if p.Key != "legacy-openai-key" || p.XToken != "legacy-x-token" || p.TextProvider != "openai" || p.ResearchProvider != "auto" || p.DeepSeekModel == "" {
		t.Fatal("legacy credentials or defaults lost")
	}
}

func TestDeepSeekRoutingKeepsNarrationAndSearchCredentialsSeparate(t *testing.T) {
	a, _, _, _ := fixture(t)
	p := a.provider()
	p.TextProvider = "deepseek"
	p.DeepSeekKey = "synthetic-deepseek"
	p.Key = "synthetic-openai"
	calls := []string{}
	a.client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls = append(calls, r.URL.Host+r.URL.Path)
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		var result []byte
		if r.URL.Host == "api.deepseek.com" {
			if r.Header.Get("Authorization") != "Bearer synthetic-deepseek" {
				t.Error("wrong DeepSeek key")
			}
			switch r.URL.Path {
			case "/responses":
				if body["model"] != p.DeepSeekModel || body["text"] == nil {
					t.Error("structured Responses contract")
				}
				result = mockResponse(`{"passed":true,"issues":[]}`)
			case "/v1/tools/search":
				if body["count"] != float64(5) || body["need_content"] != true {
					t.Error("bounded search contract")
				}
				result = []byte(`{"results":[]}`)
			default:
				t.Error("unexpected DeepSeek endpoint")
			}
		} else if r.URL.Host == "api.openai.com" && r.URL.Path == "/v1/audio/speech" {
			if r.Header.Get("Authorization") != "Bearer synthetic-openai" {
				t.Error("DeepSeek key leaked into speech")
			}
			result = make([]byte, 2048)
		} else {
			t.Error("unexpected provider destination")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(result)))}, nil
	})}
	schema := objSchema(map[string]any{"passed": map[string]any{"type": "boolean"}, "issues": arrSchema(strSchema())}, []string{"passed", "issues"})
	if _, err := a.structured(context.Background(), p, "Produce JSON", "review", schema, 500); err != nil {
		t.Fatal(err)
	}
	e := Episode{ID: apphost.ID(), Settings: Defaults()}
	if err := a.narrate(context.Background(), p, &e, Turn{Speaker: "Alex", Text: "Synthetic speech request."}); err != nil {
		t.Fatal(err)
	}
	e.Settings.People = nil
	if _, err := a.deepSeekSearch(context.Background(), p, &e, time.Now().Add(-time.Hour), time.Now()); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 4 {
		t.Fatal(calls)
	}
}

func TestAIServiceUsesActiveModelAndAccountQueue(t *testing.T) {
	a, _, _, session := fixture(t)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/_internal/chat" || r.Header.Get("Authorization") != "Bearer "+strings.Repeat("s", 40) || r.Header.Get("X-AI-App") != "podcasts" || r.Header.Get("X-AI-User") != session.User.ID {
			t.Error("AI Service app/account delegation contract")
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["model"] != nil || body["stream"] != false || body["format"] == nil {
			t.Error("active model or structured output contract")
		}
		options := body["options"].(map[string]any)
		if options["num_ctx"] != float64(32768) || options["num_predict"].(float64) > 16000 {
			t.Error("local model bounds")
		}
		io.WriteString(w, `{"done":true,"done_reason":"stop","message":{"content":"{\"passed\":true,\"issues\":[]}"}}`)
	}))
	defer server.Close()
	p := a.provider()
	p.TextProvider = "aiservice"
	p.ServiceURL = server.URL
	p.ServiceToken = strings.Repeat("s", 40)
	ctx := context.WithValue(context.Background(), generationIdentity{}, session.User.ID)
	if _, err := a.structured(ctx, p, "Return JSON", "review", map[string]any{"type": "object"}, 30000); err != nil {
		t.Fatal(err)
	}
	if _, err := a.serviceText(context.Background(), p, "test", nil, 10); err == nil {
		t.Fatal("missing account accepted")
	}
	for _, bad := range []string{"https://example.com", "http://127.0.0.1:9042/aiservice", "http://localhost:9042?x=1", "http://user:pass@127.0.0.1:9042", "http://192.168.1.2:9042"} {
		p.ServiceURL = bad
		if _, err := a.serviceText(ctx, p, "test", nil, 10); err == nil {
			t.Fatal("unsafe endpoint", bad)
		}
	}
	if requests != 1 {
		t.Fatal("invalid calls reached a server")
	}
}

func TestLocalPipelineWaitsForVoicesAndResumesWithoutRepeatingInference(t *testing.T) {
	a, h, c, session := fixture(t)
	p := a.provider()
	p.TextProvider = "aiservice"
	p.ResearchProvider = "auto"
	p.ServiceURL = "http://127.0.0.1:19552"
	p.ServiceToken = strings.Repeat("s", 40)
	p.Key = ""
	a.p = p
	settings := Defaults()
	settings.Minutes = 8
	settings.Feeds = []string{"bbc"}
	settings.People = nil
	e, err := a.create(session.User.ID, settings)
	if err != nil {
		t.Fatal(err)
	}
	localCalls, speechCalls := 0, 0
	a.client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		var content string
		if r.URL.Path == "/_internal/chat" {
			localCalls++
			if r.Header.Get("X-AI-User") != session.User.ID {
				t.Error("background identity missing")
			}
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			prompt := body["messages"].([]any)[0].(map[string]any)["content"].(string)
			if strings.Contains(prompt, "Write an original premium") {
				raw, _ := json.Marshal(sampleScript())
				content = string(raw)
			} else if strings.Contains(prompt, "podcast editor") {
				content = `{"passed":true,"issues":[]}`
			} else {
				content = strings.Repeat("Synthetic evidence S1 and S2 explain the latest reported change and a thoughtful counterpoint. ", 8)
			}
			raw, _ := json.Marshal(map[string]any{"done": true, "message": map[string]string{"content": content}})
			content = string(raw)
		} else if r.URL.Path == "/v1/audio/speech" {
			speechCalls++
			content = string(make([]byte, 2048))
		} else if r.URL.Host == "feeds.bbci.co.uk" {
			content = "<rss><channel>"
			for i := 0; i < 3; i++ {
				content += fmt.Sprintf("<item><title>Synthetic source %d</title><link>https://example.com/%d</link><pubDate>%s</pubDate><description>Synthetic evidence only.</description></item>", i, i, time.Now().UTC().Add(-time.Hour).Format(time.RFC1123Z))
			}
			content += "</channel></rss>"
		} else {
			t.Error("local mode made an unexpected cloud/search request", r.URL.Host)
			content = ""
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(content))}, nil
	})}
	if _, err = a.work(context.Background()); err != nil {
		t.Fatal(err)
	}
	result, err := a.load(e.ID, session.User.ID)
	if err != nil || result.Status != "awaiting_voice" || !result.Review.Passed || len(result.Script.Segments) == 0 || localCalls != 3 || speechCalls != 0 {
		t.Fatalf("local production %s calls=%d speech=%d err=%v", result.Status, localCalls, speechCalls, err)
	}
	// Continuing a previously reviewed script does not consume a fresh daily
	// research allowance, even if all three production slots have been used.
	for i := 0; i < 2; i++ {
		_, _ = h.DB.Exec("INSERT INTO podcast_productions VALUES(?,?)", session.User.ID, time.Now().Unix())
	}
	if w := request(h, "POST", "/api/episodes/"+e.ID+"/retry", nil, c, session.CSRF); w.Code != 409 {
		t.Fatal("resumed without a voice key")
	}
	a.p.Key = "synthetic-voice-only"
	a.p.ServiceToken = "" // Finished narration does not need the old writer grant.
	if w := request(h, "POST", "/api/episodes/"+e.ID+"/retry", nil, c, session.CSRF); w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	if _, err = a.work(context.Background()); err != nil {
		t.Fatal(err)
	}
	result, err = a.load(e.ID, session.User.ID)
	if err != nil || result.Status != "ready" || localCalls != 3 || speechCalls == 0 {
		t.Fatal("resume repeated local inference or failed", result.Status, localCalls, speechCalls, err)
	}
}

func TestDeepSeekEvidenceDatesAndUnsafeLinks(t *testing.T) {
	if _, err := parseDeepSeekSearch([]byte(`{"error":"unusable response"}`), time.Now(), time.Now()); err == nil {
		t.Fatal("accepted missing search results")
	}
	now := time.Now().UTC()
	raw, _ := json.Marshal(map[string]any{"results": []any{map[string]any{"title": "Fresh", "url": "https://example.com/new", "date": now.Add(-time.Hour).Format(time.RFC3339), "summary": "Current evidence"}, map[string]any{"title": "Old", "url": "https://example.com/old", "date": "2020-01-01T00:00:00Z"}, map[string]any{"title": "Unsafe", "url": "javascript:alert(1)"}, map[string]any{"title": "Undated context", "url": "https://example.com/context", "date": nil}}})
	ss, err := parseDeepSeekSearch(raw, now.Add(-24*time.Hour), now)
	if err != nil || len(ss) != 2 || ss[0].Title != "Fresh" || ss[1].Published != "" {
		t.Fatal(ss, err)
	}
}

func TestLocalEditorialBatchesLongScriptsAndRejectsTruncation(t *testing.T) {
	a, _, _, session := fixture(t)
	p := a.provider()
	p.TextProvider = "aiservice"
	p.ServiceURL = "http://127.0.0.1:19552"
	p.ServiceToken = strings.Repeat("s", 40)
	calls := 0
	truncated := false
	a.client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		text := body["messages"].([]any)[0].(map[string]any)["content"].(string)
		if len(text) > 48000 {
			t.Error("context budget exceeded")
		}
		if !truncated && !strings.Contains(text, reviewSpokenStyle) {
			t.Error("local review did not get the story-first narration policy")
		}
		done := `"stop"`
		if truncated {
			done = `"length"`
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"done":true,"done_reason":` + done + `,"message":{"content":"{\"passed\":true,\"issues\":[]}"}}`))}, nil
	})}
	e := Episode{Research: strings.Repeat("Synthetic research. ", 1000), Script: sampleScript(), Sources: []Source{}}
	for i := 0; i < 40; i++ {
		e.Sources = append(e.Sources, Source{ID: fmt.Sprintf("S%d", i), URL: "https://example.com/source", Title: "Synthetic source", Excerpt: strings.Repeat("evidence ", 100)})
	}
	e.Script.Segments[2].Turns = nil
	for i := 0; i < 40; i++ {
		e.Script.Segments[2].Turns = append(e.Script.Segments[2].Turns, Turn{Speaker: "Alex", Text: strings.Repeat("Synthetic claim. ", 150), Sources: []string{"S1"}})
	}
	ctx := context.WithValue(context.Background(), generationIdentity{}, session.User.ID)
	review, err := a.localEditorial(ctx, p, &e)
	if err != nil || !review.Passed || calls < 2 {
		t.Fatal("long local review", calls, err)
	}
	truncated = true
	if _, err = a.serviceText(ctx, p, "test", nil, 10); err == nil {
		t.Fatal("accepted truncated model output")
	}
}
