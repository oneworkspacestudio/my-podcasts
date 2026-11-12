package podcasts

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func fixture(t *testing.T) (*App, *apphost.Host, *http.Cookie, apphost.Session) {
	t.Helper()
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("PODCASTS_AI_SERVICE_URL", "")
	t.Setenv("PODCASTS_AI_SERVICE_TOKEN", "")
	t.Setenv("PODCASTS_X_BEARER_TOKEN", "")
	h, err := apphost.New(apphost.Config{ID: "podcasts", Name: "My Podcasts", Version: "test", Data: filepath.Join(t.TempDir(), "data"), Listen: "127.0.0.1:19444", Origin: "http://127.0.0.1:19444", RestoreFiles: []string{"providers.key"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { h.Close() })
	if err = h.Bootstrap("podcast-test@example.invalid", "Synthetic-podcast-only-2026!"); err != nil {
		t.Fatal(err)
	}
	a, err := New(h)
	if err != nil {
		t.Fatal(err)
	}
	// Existing cloud regression fixtures explicitly select the optional provider.
	a.p.SpeechProvider = "openai"
	w := request(h, "POST", "/api/login", map[string]string{"email": "podcast-test@example.invalid", "password": "Synthetic-podcast-only-2026!"}, nil, "")
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	cookie := w.Result().Cookies()[0]
	w = request(h, "GET", "/api/me", nil, cookie, "")
	var session apphost.Session
	if json.Unmarshal(w.Body.Bytes(), &session) != nil {
		t.Fatal("session")
	}
	return a, h, cookie, session
}
func request(h *apphost.Host, method, path string, body any, cookie *http.Cookie, csrf string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	r := httptest.NewRequest(method, h.Config.Origin+path, reader)
	r.RemoteAddr = "127.0.0.1:12345"
	r.Header.Set("Origin", h.Config.Origin)
	r.Header.Set("X-CSRF-Token", csrf)
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	h.Handler().ServeHTTP(w, r)
	return w
}
func decodeEpisode(t *testing.T, w *httptest.ResponseRecorder) Episode {
	t.Helper()
	var e Episode
	if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil {
		t.Fatal(err)
	}
	return e
}
func TestAccountIsolationOwnerKeysAndCSRF(t *testing.T) {
	a, h, c, s := fixture(t)
	if w := request(h, "GET", "/api/state", nil, nil, ""); w.Code != 401 {
		t.Fatal("anonymous state", w.Code)
	}
	if w := request(h, "PUT", "/api/settings", Defaults(), c, ""); w.Code != 403 {
		t.Fatal("missing CSRF")
	}
	if w := request(h, "POST", "/api/episodes", nil, c, s.CSRF); w.Code != 409 {
		t.Fatal("missing provider", w.Code)
	}
	cfg := map[string]any{"key": "synthetic-secret-not-a-real-key", "xToken": "synthetic-x-token", "model": "gpt-5.4", "ttsModel": "gpt-4o-mini-tts"}
	w := request(h, "PUT", "/api/providers", cfg, c, s.CSRF)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	w = request(h, "GET", "/api/state", nil, c, "")
	if bytes.Contains(w.Body.Bytes(), []byte("synthetic-secret")) || bytes.Contains(w.Body.Bytes(), []byte("synthetic-x-token")) {
		t.Fatal("secret returned")
	}
	var encrypted []byte
	_ = h.DB.QueryRow("SELECT secret FROM podcast_providers").Scan(&encrypted)
	if bytes.Contains(encrypted, []byte("synthetic-secret")) {
		t.Fatal("secret stored in plaintext")
	}
	if w = request(h, "POST", "/api/local-users", map[string]string{"email": "member@example.invalid", "password": "Synthetic-member-2026!"}, c, s.CSRF); w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	w = request(h, "POST", "/api/login", map[string]string{"email": "member@example.invalid", "password": "Synthetic-member-2026!"}, nil, "")
	member := w.Result().Cookies()[0]
	w = request(h, "GET", "/api/me", nil, member, "")
	var ms apphost.Session
	_ = json.Unmarshal(w.Body.Bytes(), &ms)
	if w = request(h, "PUT", "/api/providers", cfg, member, ms.CSRF); w.Code != 403 {
		t.Fatal("member provider access")
	}
	e, err := a.create(s.User.ID, Defaults())
	if err != nil {
		t.Fatal(err)
	}
	for _, endpoint := range []string{"", "/audio", "/transcript", "/sources"} {
		if w = request(h, "GET", "/api/episodes/"+e.ID+endpoint, nil, member, ""); w.Code != 404 {
			t.Fatal("cross-account read", endpoint, w.Code)
		}
	}
	for _, action := range []string{"cancel", "retry"} {
		if w = request(h, "POST", "/api/episodes/"+e.ID+"/"+action, nil, member, ms.CSRF); w.Code != 404 {
			t.Fatal("cross-account mutation")
		}
	}
}
func TestQueueLimitCancelAndRestartMigration(t *testing.T) {
	a, h, c, s := fixture(t)
	a.p.Key = "synthetic"
	var e Episode
	for i := 0; i < 3; i++ {
		w := request(h, "POST", "/api/episodes", nil, c, s.CSRF)
		if w.Code != 202 {
			t.Fatal(w.Code, w.Body.String())
		}
		e = decodeEpisode(t, w)
	}
	if w := request(h, "POST", "/api/episodes", nil, c, s.CSRF); w.Code != 409 {
		t.Fatal("queue limit")
	}
	if w := request(h, "POST", "/api/episodes/"+e.ID+"/cancel", nil, c, s.CSRF); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := a.checkpoint(ctx, &e); err == nil {
		t.Fatal("cancel ignored")
	}
	data := h.Config.Data
	_ = h.Close()
	newHost, err := apphost.New(h.Config, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer newHost.Close()
	b, err := New(newHost)
	if err != nil {
		t.Fatal("restart migration", err)
	}
	saved, err := b.load(e.ID, s.User.ID)
	if err != nil || saved.Status != "cancelled" {
		t.Fatal("restart lost state")
	}
	if _, err = os.Stat(filepath.Join(data, "providers.key")); err != nil {
		t.Fatal(err)
	}
}
func TestFeedFreshnessAndAtom(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	f := Feeds[0]
	rss := `<rss><channel><item><title>New &amp; current</title><link>https://example.com/fresh</link><pubDate>Fri, 06 Nov 2026 10:00:00 +0000</pubDate><description>&lt;b&gt;Useful&lt;/b&gt;</description></item><item><title>Old</title><link>https://example.com/old</link><pubDate>Fri, 01 Jan 2021 10:00:00 +0000</pubDate></item><item><title>Undated</title><link>https://example.com/unknown</link></item><item><title>Unsafe link</title><link>javascript:alert(1)</link><pubDate>Fri, 06 Nov 2026 10:00:00 +0000</pubDate></item></channel></rss>`
	ss, err := parseFeed([]byte(rss), f, now.Add(-24*time.Hour), now)
	if err != nil || len(ss) != 1 || ss[0].Title != "New & current" || ss[0].Excerpt != "Useful" {
		t.Fatalf("rss %#v %v", ss, err)
	}
	atom := `<feed xmlns="http://www.w3.org/2005/Atom"><entry><title>Current atom</title><link href="https://example.com/atom" rel="alternate"/><published>2026-11-06T11:00:00Z</published></entry></feed>`
	ss, err = parseFeed([]byte(atom), f, now.Add(-time.Hour), now)
	if err != nil || len(ss) != 1 {
		t.Fatal("atom", err)
	}
}
func TestScheduleUTCAndSettings(t *testing.T) {
	s := Defaults()
	s.Schedule = "daily"
	s.At = "07:00"
	now := time.Date(2026, 11, 6, 7, 0, 0, 0, time.UTC)
	if nextRun(s, now) != now.Add(24*time.Hour).Unix() {
		t.Fatal("daily boundary")
	}
	s.Schedule = "weekly"
	want := time.Date(2026, 11, 9, 7, 0, 0, 0, time.UTC)
	if nextRun(s, now) != want.Unix() {
		t.Fatal("weekly")
	}
	s.Schedule = "off"
	if nextRun(s, now) != 0 {
		t.Fatal("off")
	}
	s.People = append(s.People, "invented")
	if validSettings(s) == nil {
		t.Fatal("unverified person")
	}
}
func sampleScript() Script {
	turns := []Turn{}
	for i := 0; i < 16; i++ {
		speaker := "Alex"
		if i%2 != 0 {
			speaker = "Jamie"
		}
		turns = append(turns, Turn{Speaker: speaker, Text: strings.Repeat("Synthetic reporting gives this example enough speaking words for a complete test. ", 5), Sources: []string{"S1", "S2"}})
	}
	return Script{Title: "Synthetic technology conversation", Summary: "Synthetic production fixture; this is not a real news episode.", Segments: []Segment{{Title: "The central question", Kind: "hook", Turns: []Turn{{Speaker: "Alex", Text: "Which signal matters?", Sources: []string{}}}}, {Title: "Welcome", Kind: "intro", Turns: []Turn{{Speaker: "Jamie", Text: "We are AI-generated narrators, in a synthetic test.", Sources: []string{}}}}, {Title: "Synthetic story", Kind: "story", Turns: turns}, {Title: "What to watch", Kind: "outro", Turns: []Turn{{Speaker: "Jamie", Text: "Check the original evidence and watch what changes next.", Sources: []string{}}}}}}
}

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func mockResponse(text string) []byte {
	b, _ := json.Marshal(map[string]any{"status": "completed", "output": []any{map[string]any{"type": "message", "content": []any{map[string]any{"type": "output_text", "text": text}}}}})
	return b
}
func TestCompleteProductionResumeAndEditorialGate(t *testing.T) {
	a, h, _, s := fixture(t)
	a.p.Key = "synthetic-only"
	setting := Defaults()
	setting.Minutes = 8
	e, err := a.create(s.User.ID, setting)
	if err != nil {
		t.Fatal(err)
	}
	speech := 0
	responses := 0
	blocked := false
	mu := sync.Mutex{}
	a.client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/v1/audio/speech" {
			mu.Lock()
			speech++
			mu.Unlock()
			pcm := make([]byte, 24000)
			for i := 0; i < len(pcm)/2; i++ {
				binary.LittleEndian.PutUint16(pcm[i*2:], uint16(int16(6000*math.Sin(float64(i)*.06))))
			}
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(bytes.NewReader(pcm))}, nil
		}
		if r.URL.Path == "/v1/responses" {
			responses++
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			var b []byte
			if _, ok := body["tools"]; ok {
				b = []byte(fmt.Sprintf(`{"status":"completed","output":[{"type":"web_search_call","action":{"sources":[{"url":"https://example.com/one","title":"One"},{"url":"https://independent.example/two","title":"Two"},{"url":"https://primary.example/three","title":"Three"}]}},{"type":"message","content":[{"type":"output_text","text":%q}]}]}`, strings.Repeat("Synthetic sourced research. ", 30)))
			} else {
				format := body["text"].(map[string]any)["format"].(map[string]any)
				if format["name"] == "editorial_review" {
					review, _ := json.Marshal(Review{Passed: !blocked, Issues: []string{}})
					b = mockResponse(string(review))
				} else {
					script, _ := json.Marshal(sampleScript())
					b = mockResponse(string(script))
				}
			}
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(bytes.NewReader(b))}, nil
		}
		return &http.Response{StatusCode: 503, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("unavailable"))}, nil
	})}
	if found, err := a.work(context.Background()); err != nil || !found {
		t.Fatal(err)
	}
	result, err := a.load(e.ID, s.User.ID)
	if err != nil || result.Status != "ready" || !result.Review.Passed || result.Duration <= 0 {
		t.Fatalf("production %+v %v", result, err)
	}
	if speech < 2 || responses != 3 {
		t.Fatal("pipeline steps", speech, responses)
	}
	wav, err := os.ReadFile(filepath.Join(a.objectDir(e.ID), "episode.wav"))
	if err != nil || string(wav[:4]) != "RIFF" || string(wav[8:12]) != "WAVE" {
		t.Fatal("wave")
	}
	if int64(binary.LittleEndian.Uint32(wav[40:])) != int64(len(wav)-44) {
		t.Fatal("wav data length")
	}
	before := speech
	if err = a.produce(context.Background(), a.provider(), &result); err != nil || speech != before {
		t.Fatal("resume repeated paid speech", err, speech, before)
	}
	e2, err := a.create(s.User.ID, setting)
	if err != nil {
		t.Fatal(err)
	}
	blocked = true
	if _, err = a.work(context.Background()); err != nil {
		t.Fatal(err)
	}
	result, err = a.load(e2.ID, s.User.ID)
	if err != nil || result.Status != "needs_review" || speech != before {
		t.Fatal("failed editorial gate narrated", result.Status, err)
	}
	_ = h
}
func TestScriptEvidenceAndAudioLevels(t *testing.T) {
	s := sampleScript()
	sources := []Source{{ID: "S1"}, {ID: "S2"}}
	if _, err := validateScript(s, sources); err != nil {
		t.Fatal(err)
	}
	s.Segments[2].Turns[0].Sources = []string{"invented"}
	if _, err := validateScript(s, sources); err == nil {
		t.Fatal("invented citation")
	}
	b := make([]byte, 4000)
	for i := 0; i < 2000; i++ {
		binary.LittleEndian.PutUint16(b[i*2:], uint16(int16(32700)))
	}
	balance(b)
	for i := 0; i < 2000; i++ {
		if math.Abs(float64(int16(binary.LittleEndian.Uint16(b[i*2:])))) > 29204 {
			t.Fatal("peak ceiling")
		}
	}
}
func TestProviderErrorsDoNotExposeCredentials(t *testing.T) {
	a, _, _, _ := fixture(t)
	a.client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 401, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("private-provider-body secret-value"))}, nil
	})}
	_, err := a.post(context.Background(), providers{Key: "secret-value"}, "/responses", map[string]string{}, 1024)
	if err == nil || strings.Contains(err.Error(), "secret-value") {
		t.Fatal("unsafe provider error", err)
	}
}

func TestBudgetSurvivesDeletionAndSchedulerCatchesUpOnce(t *testing.T) {
	a, h, c, s := fixture(t)
	a.p.Key = "synthetic"
	settings := Defaults()
	settings.Schedule = "daily"
	settings.At = "07:00"
	raw, _ := json.Marshal(settings)
	if _, err := h.DB.Exec("INSERT INTO podcast_settings VALUES(?,?,?)", s.User.ID, string(raw), time.Now().Add(-72*time.Hour).Unix()); err != nil {
		t.Fatal(err)
	}
	if err := a.schedule(time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := a.schedule(time.Now()); err != nil {
		t.Fatal(err)
	}
	var count int
	_ = h.DB.QueryRow("SELECT count(*) FROM podcast_episodes WHERE user_id=?", s.User.ID).Scan(&count)
	if count != 1 {
		t.Fatal("missed schedules were replayed", count)
	}
	rows, err := h.DB.Query("SELECT body FROM podcast_episodes WHERE user_id=?", s.User.ID)
	if err != nil {
		t.Fatal(err)
	}
	var body string
	rows.Next()
	_ = rows.Scan(&body)
	rows.Close()
	var first Episode
	_ = json.Unmarshal([]byte(body), &first)
	for i := 0; i < 2; i++ {
		if _, err = a.create(s.User.ID, settings); err != nil {
			t.Fatal(err)
		}
	}
	if w := request(h, "POST", "/api/episodes/"+first.ID+"/cancel", nil, c, s.CSRF); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := request(h, "DELETE", "/api/episodes/"+first.ID, nil, c, s.CSRF); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if _, err = a.create(s.User.ID, settings); err == nil {
		t.Fatal("deleting an episode reset paid production allowance")
	}
}

func TestProductionAndDueScheduleFinishWithoutBrowserSession(t *testing.T) {
	a, h, cookie, session := fixture(t)
	user := session.User.ID
	w := request(h, "PUT", "/api/providers", map[string]string{"key": "synthetic-key"}, cookie, session.CSRF)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	settings := Defaults()
	settings.Minutes, settings.Schedule = 8, "daily"
	settings.Feeds, settings.People = []string{"bbc"}, nil
	w = request(h, "PUT", "/api/settings", settings, cookie, session.CSRF)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	now := time.Now()
	date := now.Add(-time.Hour).UTC().Format(time.RFC1123Z)
	feed := "<rss><channel>"
	for i := 0; i < 3; i++ {
		feed += fmt.Sprintf("<item><title>Synthetic source %d</title><link>https://example.com/%d</link><pubDate>%s</pubDate><description>Synthetic evidence only.</description></item>", i, i, date)
	}
	feed += "</channel></rss>"
	script, err := json.Marshal(sampleScript())
	if err != nil {
		t.Fatal(err)
	}
	recording := make(chan struct{})
	releaseRecording := make(chan struct{})
	var firstRecording sync.Once
	var searchBudgets []int
	a.client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		var result []byte
		switch r.URL.Path {
		case "/v1/audio/speech":
			firstRecording.Do(func() { close(recording) })
			select {
			case <-releaseRecording:
			case <-r.Context().Done():
				return nil, r.Context().Err()
			}
			result = make([]byte, 2048)
		case "/v1/responses":
			var body struct {
				Tools []any  `json:"tools"`
				Max   int    `json:"max_output_tokens"`
				Input string `json:"input"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				return nil, err
			}
			switch {
			case len(body.Tools) > 0:
				searchBudgets = append(searchBudgets, body.Max)
				result = []byte(`{"status":"incomplete","incomplete_details":{"reason":"max_output_tokens"}}`)
			case strings.Contains(body.Input, "Write an original premium"):
				if !strings.Contains(body.Input, "Synthetic source 0") {
					t.Error("writer did not get the dated fallback evidence")
				}
				result = mockResponse(string(script))
			default:
				result = mockResponse(`{"passed":true,"issues":[]}`)
			}
		default:
			if r.URL.Host != "feeds.bbci.co.uk" {
				return nil, fmt.Errorf("unexpected provider request to %s", r.URL.Host)
			}
			result = []byte(feed)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(bytes.NewReader(result))}, nil
	})}
	// Enqueue through the normal authenticated endpoint before starting the
	// supervised loop. Its provider requests must outlive the HTTP request.
	w = request(h, "POST", "/api/episodes", nil, cookie, session.CSRF)
	if w.Code != 202 {
		t.Fatal(w.Code, w.Body.String())
	}
	manual := decodeEpisode(t, w)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(5 * time.Second):
			t.Error("server worker did not stop")
		}
	})
	select {
	case <-recording:
	case <-time.After(5 * time.Second):
		t.Fatal("manual episode did not reach server-side narration")
	}
	due := now.Add(-time.Minute).Unix()
	if _, err := h.DB.Exec("UPDATE podcast_settings SET next_run=? WHERE user_id=?", due, user); err != nil {
		t.Fatal(err)
	}
	settings.ShowName = "Updated show while a run is due"
	w = request(h, "PUT", "/api/settings", settings, cookie, session.CSRF)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if err := a.schedule(time.Now()); err != nil {
		t.Fatal(err)
	}
	var next int64
	if err := h.DB.QueryRow("SELECT next_run FROM podcast_settings WHERE user_id=?", user).Scan(&next); err != nil || next != due {
		t.Fatalf("due schedule moved while another production was recording: %d, %v", next, err)
	}
	w = request(h, "POST", "/auth/logout", nil, cookie, session.CSRF)
	if w.Code != 200 || request(h, "GET", "/api/state", nil, cookie, "").Code != 401 {
		t.Fatalf("synthetic session was not closed: %d", w.Code)
	}
	close(releaseRecording)
	var ready int
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); {
		if err := h.DB.QueryRow("SELECT count(*) FROM podcast_episodes WHERE user_id=? AND status='ready'", user).Scan(&ready); err != nil {
			t.Fatal(err)
		}
		if ready == 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if ready != 2 {
		t.Fatal("manual and automatic productions did not finish without a session", ready)
	}
	w = request(h, "POST", "/api/login", map[string]string{"email": "podcast-test@example.invalid", "password": "Synthetic-podcast-only-2026!"}, nil, "")
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	newCookie := w.Result().Cookies()[0]
	var state struct {
		Episodes []Episode
		NextRun  int64
	}
	w = request(h, "GET", "/api/state", nil, newCookie, "")
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil || len(state.Episodes) != 2 || state.NextRun <= time.Now().Unix() {
		t.Fatalf("ready episodes and next schedule not available on return: %+v, %v", state, err)
	}
	for _, episode := range state.Episodes {
		w = request(h, "GET", "/api/episodes/"+episode.ID+"/audio", nil, newCookie, "")
		if w.Code != 200 || !bytes.HasPrefix(w.Body.Bytes(), []byte("RIFF")) || episode.ResearchProvider != "feeds" || len(episode.Sources) != 3 {
			t.Fatalf("ready episode did not have playable audio and dated evidence: %d", w.Code)
		}
		if episode.ID != manual.ID && episode.Settings.ShowName != settings.ShowName {
			t.Fatal("scheduled episode did not use the latest saved show settings")
		}
	}
	if fmt.Sprint(searchBudgets) != "[16000 32000 16000 32000]" {
		t.Fatal("research retries were not bounded", searchBudgets)
	}
}

func TestCancelActiveProviderRequestAndStartupRecovery(t *testing.T) {
	a, h, c, s := fixture(t)
	a.p.Key = "synthetic"
	setting := Defaults()
	setting.Minutes = 8
	e, err := a.create(s.User.ID, setting)
	if err != nil {
		t.Fatal(err)
	}
	e.Research = "Synthetic saved research"
	e.Script = sampleScript()
	e.Review = Review{Passed: true, Issues: []string{}}
	e.Sources = []Source{{ID: "S1"}, {ID: "S2"}}
	if err = a.save(&e); err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	a.client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		close(entered)
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	done := make(chan error, 1)
	go func() { _, err := a.work(context.Background()); done <- err }()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("provider request did not start")
	}
	if w := request(h, "POST", "/api/episodes/"+e.ID+"/cancel", nil, c, s.CSRF); w.Code != 200 {
		t.Fatal(w.Code)
	}
	select {
	case err = <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancel did not interrupt provider")
	}
	saved, err := a.load(e.ID, s.User.ID)
	if err != nil || saved.Status != "cancelled" {
		t.Fatal("cancel overwritten")
	}
	// Simulate a crash at a saved stage, then exercise the real unattended Run loop.
	saved.Status = "running"
	if err = a.save(&saved); err != nil {
		t.Fatal(err)
	}
	a.p.Key = ""
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { done <- a.Run(ctx) }()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		saved, err = a.load(e.ID, s.User.ID)
		if err == nil && saved.Status == "awaiting_voice" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if saved.Status != "awaiting_voice" || !strings.Contains(saved.Error, "voices") {
		t.Fatal("startup did not resume persisted production", saved.Status)
	}
	cancel()
	if err = <-done; err != nil {
		t.Fatal(err)
	}
}

func TestLivePublicFeeds(t *testing.T) {
	if os.Getenv("PODCASTS_LIVE_SOURCES") != "1" {
		t.Skip("explicit network acceptance only")
	}
	a, _, _, _ := fixture(t)
	now := time.Now().UTC()
	var wg sync.WaitGroup
	var mu sync.Mutex
	report := []map[string]any{}
	for _, feed := range Feeds {
		wg.Add(1)
		go func(f Feed) {
			defer wg.Done()
			b, err := a.get(context.Background(), f.URL, "", 3<<20)
			n := 0
			state := "available"
			if err == nil {
				var sources []Source
				sources, err = parseFeed(b, f, now.Add(-7*24*time.Hour), now)
				n = len(sources)
			}
			if err != nil {
				state = err.Error()
			}
			mu.Lock()
			report = append(report, map[string]any{"name": f.Name, "url": f.URL, "status": state, "datedStoriesPastWeek": n})
			mu.Unlock()
		}(feed)
	}
	wg.Wait()
	raw, _ := json.MarshalIndent(map[string]any{"checkedAt": now.Format(time.RFC3339), "scope": "real public RSS availability; not paid AI or credentialed socials", "feeds": report}, "", "  ")
	if err := os.WriteFile(filepath.Join("..", "..", "docs", "evidence", "live-feeds.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	available := 0
	for _, r := range report {
		if r["status"] == "available" {
			available++
		}
	}
	if available == 0 {
		t.Fatal("all public sources unavailable")
	}
}
