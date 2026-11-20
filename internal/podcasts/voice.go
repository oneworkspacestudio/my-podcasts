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
	"slices"
	"strconv"
	"time"
)

var errVoiceUnavailable = errors.New("My Voice is temporarily unavailable")

type voiceStatus struct {
	Protocol int
	Model    string
	Cast     string
	Issue    string
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
	r, err := http.NewRequestWithContext(ctx, method, p.VoiceURL+path, bytes.NewReader(body))
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
	res, err := a.voiceRequest(ctx, p, "GET", "/v2/audio/models", nil)
	if err != nil {
		return voiceStatus{Issue: "My Voice is offline. Start it on this Workspace server"}
	}
	if res.StatusCode != http.StatusNotFound {
		defer res.Body.Close()
		if res.StatusCode == http.StatusUnauthorized {
			return voiceStatus{Issue: "My Voice rejected the API key. Create a key in My Voice Settings"}
		}
		var catalog struct {
			Protocol int
			Choices  []struct {
				Value     string
				Installed bool
				Voices    []string
			}
		}
		if res.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(res.Body, 32768)).Decode(&catalog) != nil || catalog.Protocol != 2 {
			return voiceStatus{Issue: "Could not read the My Voice model list"}
		}
		order := []string{"pocket18", "pocket7", "moss", "qwen06", "qwen17"}
		if p.LocalVoiceModel != "" {
			order = []string{p.LocalVoiceModel}
		}
		for _, duo := range []bool{true, false} {
			for _, id := range order {
				for _, choice := range catalog.Choices {
					if choice.Value != id || !choice.Installed || !slices.Contains(choice.Voices, "default") {
						continue
					}
					cast := "single-v1"
					if slices.Contains(choice.Voices, "alba") && slices.Contains(choice.Voices, "marius") {
						cast = "marius-alba-v1"
					}
					if (cast == "marius-alba-v1") == duo {
						return voiceStatus{Protocol: 2, Model: id, Cast: cast}
					}
				}
			}
		}
		return voiceStatus{Issue: "Install the selected English narration model in My Voice Settings"}
	}
	res.Body.Close()
	if p.LocalVoiceModel != "" {
		return voiceStatus{Issue: "Update My Voice to use the narration model selected in My Podcasts"}
	}
	res, err = a.voiceRequest(ctx, p, "GET", "/_internal/podcasts/status", nil)
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
	if p.VoiceURL == a.p.VoiceURL && p.VoiceToken == a.p.VoiceToken && p.LocalVoiceModel == a.p.LocalVoiceModel {
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
	var turn map[string]any
	path := "/_internal/podcasts/speech"
	if p.voiceProtocol == 2 {
		path = "/v2/audio/speech"
		voice, speed, pitch := "default", 1.02, 1.0
		if t.Speaker == "Jamie" {
			speed, pitch = 0.96, -1.5
		}
		if p.voiceCast == "marius-alba-v1" {
			voice, speed, pitch = "marius", 1, 0
			if t.Speaker == "Jamie" {
				voice = "alba"
			}
		}
		turn = map[string]any{"input": cleanTranscript(t.Text), "model": p.TTSModel, "voice": voice, "response_format": "pcm", "speed": speed, "pitch": pitch}
	} else {
		turn = map[string]any{"text": cleanTranscript(t.Text), "speaker": t.Speaker, "model": p.TTSModel}
		if p.voiceCast != "" {
			turn["cast"] = p.voiceCast
		}
	}
	body, err := json.Marshal(turn)
	if err != nil {
		return nil, err
	}
	res, err := a.voiceRequest(ctx, p, "POST", path, body)
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
