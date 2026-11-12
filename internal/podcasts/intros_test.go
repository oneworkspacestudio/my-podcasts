package podcasts

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func introTestPCM(seconds int, hz float64) []byte {
	b := make([]byte, seconds*sampleRate*2)
	for i := range seconds * sampleRate {
		binary.LittleEndian.PutUint16(b[i*2:], uint16(int16(5000*math.Sin(2*math.Pi*hz*float64(i)/sampleRate))))
	}
	return b
}

func TestIntroTracksPinnedAndMasteredAfterHook(t *testing.T) {
	a, h, cookie, session := fixture(t)
	first, second := introTestPCM(1, 220), introTestPCM(2, 330)
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	for i, field := range []string{"first", "second"} {
		part, err := form.CreateFormFile(field, field+".wav")
		if err != nil {
			t.Fatal(err)
		}
		pcm := [][]byte{first, second}[i]
		_, _ = part.Write(append(wavHeader(uint32(len(pcm))), pcm...))
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	upload := func(csrf string) int {
		req := httptest.NewRequest("PUT", h.Config.Origin+"/api/intros", bytes.NewReader(body.Bytes()))
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Content-Type", form.FormDataContentType())
		req.Header.Set("Origin", h.Config.Origin)
		req.Header.Set("X-CSRF-Token", csrf)
		req.AddCookie(cookie)
		w := httptest.NewRecorder()
		h.Handler().ServeHTTP(w, req)
		return w.Code
	}
	if got := upload(""); got != http.StatusForbidden {
		t.Fatalf("upload without CSRF: %d", got)
	}
	if got := upload(session.CSRF); got != 200 {
		t.Fatalf("upload: %d", got)
	}
	e, err := a.create(session.User.ID, Defaults())
	if err != nil {
		t.Fatal(err)
	}
	expected := first
	if strings.ContainsRune("13579bdf", rune(e.ID[0])) {
		expected = second
	}
	if w := request(h, "DELETE", "/api/intros", nil, cookie, session.CSRF); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	next, err := a.create(session.User.ID, Defaults())
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := h.DB.QueryRow("SELECT count(*) FROM podcast_episode_intros WHERE episode_id=?", next.ID).Scan(&count); err != nil || count != 0 {
		t.Fatal("a disabled intro was queued", count, err)
	}
	e.Script = sampleScript()
	e.Script.Segments = e.Script.Segments[:2]
	addHostIntroductions(&e.Script)
	if got := e.Script.Segments[1].Turns; len(got) != 2 || got[0].Speaker != "Alex" || got[0].Text != "I’m Alex." ||
		got[1].Speaker != "Jamie" || got[1].Text != "Hi, I’m Jamie. Let’s start with the signal that changed this week." {
		t.Fatalf("the hosts did not introduce themselves in their own turns: %+v", got)
	}
	p := a.provider()
	turnPCM := introTestPCM(1, 660)
	for _, segment := range e.Script.Segments {
		for _, turn := range segment.Turns {
			if err := a.writeObject(context.Background(), e.ID, chunkName(p, e.Settings, turn), turnPCM); err != nil {
				t.Fatal(err)
			}
		}
	}
	balance(expected)
	hookEnd := len(turnPCM) + sampleRate*2*17/100 + sampleRate
	for range 2 {
		if err := a.master(context.Background(), p, &e); err != nil {
			t.Fatal(err)
		}
		wav, err := os.ReadFile(filepath.Join(a.objectDir(e.ID), "episode.wav"))
		if err != nil || !bytes.Equal(wav[44+hookEnd:44+hookEnd+len(expected)], expected) {
			t.Fatal("the pinned track did not play directly after the hook", err)
		}
		if got, want := e.Script.Segments[1].Start, float64(hookEnd+len(expected))/(sampleRate*2); got != want {
			t.Fatalf("welcome chapter started at %f, want %f", got, want)
		}
	}
	e.Status = "ready"
	if err := a.save(&e); err != nil {
		t.Fatal(err)
	}
	if w := request(h, "DELETE", "/api/episodes/"+e.ID, nil, cookie, session.CSRF); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if err := h.DB.QueryRow("SELECT count(*) FROM podcast_intro_tracks").Scan(&count); err != nil || count != 0 {
		t.Fatal("unused tracks were not pruned", count, err)
	}
}

func TestIntroductionKeepsExistingSelfIdentification(t *testing.T) {
	s := Script{Segments: []Segment{{Kind: "intro", Turns: []Turn{
		{Speaker: "Alex", Text: "Welcome, I'm Alex."},
		{Speaker: "Jamie", Text: "I am Jamie, and the first story is a surprise."},
	}}}}
	addHostIntroductions(&s)
	if turns := s.Segments[0].Turns; turns[0].Text != "Welcome, I'm Alex." || turns[1].Text != "I am Jamie, and the first story is a surprise." {
		t.Fatalf("duplicated self-identification: %+v", turns)
	}
}
