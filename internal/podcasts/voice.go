package podcasts

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var errVoiceUnavailable = errors.New("My Voice is temporarily unavailable")

type voiceStatus struct {
	Model string
	Cast  string
	Issue string
}

func defaultVoiceURL(listen string, managed bool) string {
	// In the managed catalog Podcasts uses manager+54 and My Voice uses +38.
	if managed {
		_, port, _ := net.SplitHostPort(listen)
		n, _ := strconv.Atoi(port)
		if n > 16 && n <= 65535 {
			return fmt.Sprintf("http://127.0.0.1:%d", n-16)
		}
	}
	return "http://127.0.0.1:9038"
}

func validVoiceURL(raw string) bool {
	if !validServiceURL(raw) {
		return false
	}
	u, _ := url.Parse(raw)
	return net.ParseIP(u.Hostname()).IsLoopback()
}

func validVoiceToken(token string) bool {
	if len(token) != 64 {
		return false
	}
	_, err := hex.DecodeString(token)
	return err == nil
}

func (a *App) voiceRequest(ctx context.Context, p providers, method, path string, body []byte) (*http.Response, error) {
	if !validVoiceURL(p.VoiceURL) || !validVoiceToken(p.VoiceToken) {
		return nil, errors.New("Connect My Voice in Connections to record this episode")
	}
	r, err := http.NewRequestWithContext(ctx, method, p.VoiceURL+"/_internal/podcasts/"+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Authorization", "Bearer "+p.VoiceToken)
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
	}
	return a.voiceClient.Do(r)
}

func (a *App) probeVoice(ctx context.Context, p providers) voiceStatus {
	if !validVoiceToken(p.VoiceToken) {
		return voiceStatus{Issue: "Connect My Voice in Connections to record this episode"}
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	res, err := a.voiceRequest(ctx, p, "GET", "status", nil)
	if err != nil {
		return voiceStatus{Issue: "My Voice is offline. Start it on this Workspace server"}
	}
	defer res.Body.Close()
	if res.StatusCode == 401 || res.StatusCode == 403 {
		return voiceStatus{Issue: "My Voice rejected the connection. Issue a new Podcasts credential"}
	}
	var status struct {
		Protocol int
		Ready    bool
		Model    string
		Cast     string
	}
	if res.StatusCode != 200 || json.NewDecoder(io.LimitReader(res.Body, 4096)).Decode(&status) != nil || status.Protocol != 1 {
		return voiceStatus{Issue: "Update My Voice to a version that supports local Podcasts narration"}
	}
	if !status.Ready || !safeModel.MatchString(status.Model) {
		return voiceStatus{Issue: "Install an English speech model in My Voice Settings"}
	}
	if !safeModel.MatchString(status.Cast) {
		status.Cast = ""
	}
	return voiceStatus{Model: status.Model, Cast: status.Cast}
}

func (a *App) refreshVoice(ctx context.Context) {
	p := a.provider()
	if p.SpeechProvider != "myvoice" {
		return
	}
	status := a.probeVoice(ctx, p)
	a.mu.Lock()
	defer a.mu.Unlock()
	if p.VoiceURL == a.p.VoiceURL && p.VoiceToken == a.p.VoiceToken {
		a.voice = status
	}
}

func (a *App) speechIssueLocked(p providers) string {
	if p.SpeechProvider == "openai" {
		if p.Key == "" {
			return "Connect OpenAI speech or select My Voice in Connections"
		}
		return ""
	}
	if a.voice.Model != "" {
		return ""
	}
	if a.voice.Issue != "" {
		return a.voice.Issue
	}
	if p.VoiceToken != "" {
		return "Checking the My Voice connection"
	}
	return "Connect My Voice in Connections to record this episode"
}

func (a *App) localSpeech(ctx context.Context, p providers, t Turn) ([]byte, error) {
	turn := map[string]string{"text": cleanTranscript(t.Text), "speaker": t.Speaker, "model": p.TTSModel}
	if p.voiceCast != "" {
		turn["cast"] = p.voiceCast
	}
	body, err := json.Marshal(turn)
	if err != nil {
		return nil, err
	}
	res, err := a.voiceRequest(ctx, p, "POST", "speech", body)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errVoiceUnavailable
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 401, 403, 404, 409, 429, 503:
		return nil, errVoiceUnavailable
	case 200:
	default:
		return nil, fmt.Errorf("My Voice could not render this turn (HTTP %d). Check its model and resume", res.StatusCode)
	}
	if res.Header.Get("X-Voice-Sample-Rate") != "24000" {
		return nil, errors.New("My Voice returned an incompatible sample rate")
	}
	const limit = 16 << 20
	data, err := io.ReadAll(io.LimitReader(res.Body, limit+1))
	if err != nil || len(data) > limit {
		return nil, errors.New("My Voice returned incomplete or oversized audio")
	}
	return data, nil
}
