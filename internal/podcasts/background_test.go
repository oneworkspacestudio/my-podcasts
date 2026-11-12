package podcasts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestThreeDailyEditionsCompleteAsWAVWithoutSession(t *testing.T) {
	a, h, cookie, session := fixture(t)
	var recordings atomic.Int32
	voiceURL, voiceToken := os.Getenv("PODCASTS_TEST_VOICE_URL"), os.Getenv("PODCASTS_TEST_VOICE_TOKEN")
	realVoice := voiceURL != ""
	if !realVoice {
		voiceToken = strings.Repeat("a", 64)
		voice := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer "+voiceToken {
				w.WriteHeader(401)
				return
			}
			switch r.URL.Path {
			case "/_internal/podcasts/status":
				io.WriteString(w, `{"protocol":1,"ready":true,"model":"pocket18"}`)
			case "/_internal/podcasts/speech":
				recordings.Add(1)
				w.Header().Set("X-Voice-Sample-Rate", "24000")
				w.Write(bytes.Repeat([]byte{0x10, 0x02}, 4096))
			default:
				w.WriteHeader(404)
			}
		}))
		defer voice.Close()
		voiceURL = voice.URL
	}
	w := request(h, "PUT", "/api/providers", map[string]string{"textProvider": "deepseek", "deepSeekKey": "synthetic", "researchProvider": "feeds", "speechProvider": "myvoice", "voiceURL": voiceURL, "voiceToken": voiceToken}, cookie, session.CSRF)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	settings := Defaults()
	settings.Schedule, settings.At, settings.Minutes = "daily-three", "15:00", 8
	settings.Feeds, settings.People = []string{"bbc"}, nil
	w = request(h, "PUT", "/api/settings", settings, cookie, session.CSRF)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := request(h, "POST", "/auth/logout", nil, cookie, session.CSRF); w.Code != 200 {
		t.Fatal("logout", w.Code)
	}
	if w := request(h, "GET", "/api/state", nil, cookie, ""); w.Code != 401 {
		t.Fatal("test session still signed in", w.Code)
	}

	first := time.Now().UTC().Truncate(24 * time.Hour).Add(31 * time.Hour)
	waiting, err := a.create(session.User.ID, settings)
	if err != nil {
		t.Fatal(err)
	}
	waiting.Status, waiting.Stage, waiting.Error = "awaiting_voice", "Script ready", "Connect an OpenAI API key for voices and resume"
	waiting.Research = "Saved synthetic research"
	waiting.Script = sampleScript()
	waiting.Script.Title = "Previously reviewed script"
	waiting.Script.Segments[2].Turns = waiting.Script.Segments[2].Turns[:3]
	waiting.Review.Passed = true
	if err := a.save(&waiting); err != nil {
		t.Fatal(err)
	}
	// The previously requested script is from a prior day in the simulated clock.
	if _, err := h.DB.Exec("UPDATE podcast_productions SET started=? WHERE user_id=?", first.Add(-48*time.Hour).Unix(), session.User.ID); err != nil {
		t.Fatal(err)
	}
	var clock atomic.Int64
	clock.Store(first.Unix())
	ticks := make(chan time.Time, 1)
	date := time.Now().UTC().Add(-time.Hour).Format(time.RFC1123Z)
	feed := "<rss><channel>"
	for i := 0; i < 3; i++ {
		feed += fmt.Sprintf("<item><title>Synthetic report %d</title><link>https://example.com/%d</link><pubDate>%s</pubDate><description>Synthetic evidence.</description></item>", i, i, date)
	}
	feed += "</channel></rss>"
	var drafts, reviews atomic.Int32
	a.client = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		var body []byte
		switch {
		case r.URL.Host == "feeds.bbci.co.uk":
			body = []byte(feed)
		case r.URL.Host == "api.deepseek.com" && r.URL.Path == "/responses":
			var payload struct{ Input string }
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				return nil, err
			}
			switch {
			case strings.Contains(payload.Input, "Write an original premium"):
				drafts.Add(1)
				edition := int((clock.Load()-first.Unix())/int64(threeDailySpacing/time.Second)) + 1
				if !strings.Contains(payload.Input, fmt.Sprintf("edition %d of 3", edition)) {
					t.Errorf("edition %d is missing its editorial direction", edition)
				}
				s := sampleScript()
				s.Title = fmt.Sprintf("Synthetic edition %d", edition)
				rewrite := strings.Contains(payload.Input, "The previous draft had")
				switch {
				case edition == 1 && rewrite:
					s.Segments[2].Turns = s.Segments[2].Turns[:5]
				case edition == 1:
					s.Segments[2].Turns = s.Segments[2].Turns[:3]
				case edition == 2 && !rewrite:
					story, outro := s.Segments[2], s.Segments[3]
					s.Segments = s.Segments[:2]
					for i := 0; i < 9; i++ {
						s.Segments = append(s.Segments, story)
					}
					s.Segments = append(s.Segments, outro)
				}
				raw, err := json.Marshal(s)
				if err != nil {
					return nil, err
				}
				body = mockResponse(string(raw))
			case strings.Contains(payload.Input, "independent podcast editor"):
				reviews.Add(1)
				body = mockResponse(`{"passed":true,"issues":[]}`)
			case strings.Contains(payload.Input, "research dossier"):
				body = mockResponse(strings.Repeat("S1 and S2 report synthetic evidence. S3 provides context and a different perspective. ", 6))
			default:
				return nil, fmt.Errorf("unexpected Responses request")
			}
		default:
			return nil, fmt.Errorf("unexpected provider request")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(bytes.NewReader(body))}, nil
	})}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- a.run(ctx, func() time.Time { return time.Unix(clock.Load(), 0) }, ticks)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(5 * time.Second):
			t.Error("background worker did not stop")
		}
	})
	for edition := 1; edition <= 3; edition++ {
		if edition > 1 {
			slot := first.Add(time.Duration(edition-1) * threeDailySpacing)
			clock.Store(slot.Unix())
			ticks <- slot
		}
		var ready int
		timeout := 8 * time.Second
		if realVoice {
			timeout = 4 * time.Minute
		}
		for deadline := time.Now().Add(timeout); time.Now().Before(deadline); {
			if err := h.DB.QueryRow("SELECT count(*) FROM podcast_episodes WHERE status='ready' AND user_id=? AND id<>?", session.User.ID, waiting.ID).Scan(&ready); err != nil {
				t.Fatal(err)
			}
			if ready >= edition {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if ready != edition {
			t.Fatalf("expected %d ready episodes during sign-out, got %d", edition, ready)
		}
	}

	w = request(h, "POST", "/api/login", map[string]string{"email": "podcast-test@example.invalid", "password": "Synthetic-podcast-only-2026!"}, nil, "")
	if w.Code != 200 {
		t.Fatal("sign in after production", w.Code)
	}
	cookie = w.Result().Cookies()[0]
	var state struct {
		Episodes []Episode
		NextRun  int64
	}
	w = request(h, "GET", "/api/state", nil, cookie, "")
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil || len(state.Episodes) != 4 || state.NextRun != first.Add(24*time.Hour).Unix() {
		t.Fatalf("expected three episodes and tomorrow's next slot, got %d and %d (%v)", len(state.Episodes), state.NextRun, err)
	}
	old := state.Episodes[3]
	audio := request(h, "GET", "/api/episodes/"+waiting.ID+"/audio", nil, cookie, "")
	if old.ID != waiting.ID || old.Status != "ready" || old.Script.Title != waiting.Script.Title ||
		audio.Code != 200 || !bytes.HasPrefix(audio.Body.Bytes(), []byte("RIFF")) {
		t.Fatal("the reviewed script did not resume with local speech")
	}
	for i, episode := range state.Episodes[:3] {
		edition := 3 - i
		slot := first.Add(time.Duration(edition-1) * threeDailySpacing).Unix()
		if episode.ScheduledFor != slot || episode.Status != "ready" || !episode.Review.Passed ||
			episode.Script.Title != fmt.Sprintf("Synthetic edition %d", edition) || episode.Duration <= 0 {
			t.Fatalf("edition %d missing its slot, script, editorial pass or audio", edition)
		}
		if (edition == 1) != strings.Contains(episode.LengthNote, "shorter") {
			t.Fatalf("edition %d length note = %q", edition, episode.LengthNote)
		}
		audio := request(h, "GET", "/api/episodes/"+episode.ID+"/audio", nil, cookie, "")
		if audio.Code != 200 || audio.Body.Len() <= 44 || !bytes.HasPrefix(audio.Body.Bytes(), []byte("RIFF")) {
			t.Fatalf("edition %d did not return playable WAV audio", edition)
		}
	}
	if drafts.Load() != 5 || reviews.Load() != 3 || (!realVoice && recordings.Load() < 3) {
		t.Fatalf("expected five drafts and three editorial passes before speech, got %d/%d/%d", drafts.Load(), reviews.Load(), recordings.Load())
	}
	t.Logf("Three automatic episodes returned WAV without OpenAI speech or a signed-in browser (native My Voice: %t)", realVoice)
}
