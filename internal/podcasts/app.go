package podcasts

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const schema = `CREATE TABLE podcast_settings(user_id TEXT PRIMARY KEY,body TEXT NOT NULL,next_run INTEGER NOT NULL DEFAULT 0);
CREATE TABLE podcast_episodes(id TEXT PRIMARY KEY,user_id TEXT NOT NULL,created INTEGER NOT NULL,updated INTEGER NOT NULL,status TEXT NOT NULL,body TEXT NOT NULL);
CREATE INDEX podcast_queue ON podcast_episodes(status,created);
CREATE INDEX podcast_user ON podcast_episodes(user_id,created);
CREATE TABLE podcast_productions(user_id TEXT NOT NULL,started INTEGER NOT NULL);
CREATE INDEX podcast_budget ON podcast_productions(user_id,started);
CREATE TABLE podcast_providers(id INTEGER PRIMARY KEY CHECK(id=1),secret BLOB NOT NULL);`

type providers struct {
	TextProvider     string `json:"textProvider"`
	ResearchProvider string `json:"researchProvider"`
	DeepSeekKey      string `json:"deepSeekKey"`
	DeepSeekModel    string `json:"deepSeekModel"`
	ServiceURL       string `json:"serviceURL"`
	ServiceToken     string `json:"serviceToken"`
	Key              string `json:"key"`
	XToken           string `json:"xToken"`
	Model            string `json:"model"`
	TTSModel         string `json:"ttsModel"`
	SpeechProvider   string `json:"speechProvider"`
	VoiceURL         string `json:"voiceURL"`
	VoiceToken       string `json:"voiceToken"`
}
type App struct {
	h           *apphost.Host
	client      *http.Client
	mu          sync.Mutex
	p           providers
	key         []byte
	wake        chan struct{}
	active      map[string]context.CancelFunc
	apiBase     string
	voiceClient *http.Client
	voice       voiceStatus
}

func New(h *apphost.Host) (*App, error) {
	if e := h.Migrate(1, schema); e != nil {
		return nil, e
	}
	a := &App{h: h, client: &http.Client{Timeout: 150 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return errors.New("redirect rejected") }}, wake: make(chan struct{}, 1), active: map[string]context.CancelFunc{}, apiBase: "https://api.openai.com/v1", p: providers{DeepSeekKey: os.Getenv("DEEPSEEK_API_KEY"), ServiceToken: os.Getenv("PODCASTS_AI_SERVICE_TOKEN"), Key: os.Getenv("OPENAI_API_KEY"), XToken: os.Getenv("PODCASTS_X_BEARER_TOKEN"), Model: "gpt-5.4", TTSModel: "gpt-4o-mini-tts"}}
	keyPath := filepath.Join(h.Config.Data, "providers.key")
	k, e := os.ReadFile(keyPath)
	if os.IsNotExist(e) {
		k = make([]byte, 32)
		if _, e = rand.Read(k); e == nil {
			e = os.WriteFile(keyPath, k, 0600)
		}
	}
	if e != nil || len(k) != 32 {
		return nil, errors.New("Podcast provider encryption key unavailable")
	}
	a.key = k
	h.BackupPaths = append(h.BackupPaths, "providers.key")
	var enc []byte
	if e = h.DB.QueryRow("SELECT secret FROM podcast_providers WHERE id=1").Scan(&enc); e == nil {
		raw, e := a.crypt(enc, false)
		if e != nil || json.Unmarshal(raw, &a.p) != nil {
			return nil, errors.New("Podcast provider settings could not be decrypted")
		}
	} else if e != sql.ErrNoRows {
		return nil, e
	}
	normalizeProviders(&a.p)
	if a.p.VoiceURL == "" {
		a.p.VoiceURL = defaultVoiceURL(h.Config.Listen, h.Config.Issuer != "")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	a.voiceClient = &http.Client{Timeout: 4*time.Minute + 5*time.Second, Transport: transport, CheckRedirect: a.client.CheckRedirect}
	if !validServiceURL(a.p.ServiceURL) {
		return nil, errors.New("Configure a loopback AI Service URL with an explicit port")
	}
	h.Mux.HandleFunc("POST /api/providers/test", h.Owner(a.testConnection))
	h.Mux.HandleFunc("GET /api/state", h.Require(a.state))
	h.Mux.HandleFunc("PUT /api/settings", h.Require(a.settings))
	h.Mux.HandleFunc("PUT /api/providers", h.Owner(a.setProviders))
	h.Mux.HandleFunc("POST /api/episodes", h.Require(a.enqueue))
	h.Mux.HandleFunc("GET /api/episodes/{id}", h.Require(a.getEpisode))
	h.Mux.HandleFunc("POST /api/episodes/{id}/cancel", h.Require(a.cancel))
	h.Mux.HandleFunc("POST /api/episodes/{id}/retry", h.Require(a.retry))
	h.Mux.HandleFunc("DELETE /api/episodes/{id}", h.Require(a.deleteEpisode))
	h.Mux.HandleFunc("GET /api/episodes/{id}/audio", h.Require(a.audio))
	h.Mux.HandleFunc("GET /api/episodes/{id}/transcript", h.Require(a.transcript))
	h.Mux.HandleFunc("GET /api/episodes/{id}/sources", h.Require(a.sourceExport))
	return a, nil
}
func (a *App) crypt(b []byte, encrypt bool) ([]byte, error) {
	block, e := aes.NewCipher(a.key)
	if e != nil {
		return nil, e
	}
	g, e := cipher.NewGCM(block)
	if e != nil {
		return nil, e
	}
	if encrypt {
		n := make([]byte, g.NonceSize())
		if _, e = rand.Read(n); e != nil {
			return nil, e
		}
		return g.Seal(n, n, b, []byte("my-podcasts/providers/v1")), nil
	}
	if len(b) < g.NonceSize() {
		return nil, errors.New("invalid provider record")
	}
	return g.Open(nil, b[:g.NonceSize()], b[g.NonceSize():], []byte("my-podcasts/providers/v1"))
}
func (a *App) provider() providers { a.mu.Lock(); defer a.mu.Unlock(); return a.p }
func (a *App) setProviders(w http.ResponseWriter, r *http.Request) {
	var in struct {
		TextProvider     string `json:"textProvider"`
		ResearchProvider string `json:"researchProvider"`
		DeepSeekKey      string `json:"deepSeekKey"`
		DeepSeekModel    string `json:"deepSeekModel"`
		ServiceURL       string `json:"serviceURL"`
		ServiceToken     string `json:"serviceToken"`
		ClearDeepSeek    bool   `json:"clearDeepSeek"`
		ClearService     bool   `json:"clearService"`
		Key              string `json:"key"`
		XToken           string `json:"xToken"`
		Model            string `json:"model"`
		TTSModel         string `json:"ttsModel"`
		ClearKey         bool   `json:"clearKey"`
		ClearX           bool   `json:"clearX"`
		SpeechProvider   string `json:"speechProvider"`
		VoiceURL         string `json:"voiceURL"`
		VoiceToken       string `json:"voiceToken"`
		ClearVoice       bool   `json:"clearVoice"`
	}
	if !apphost.Decode(w, r, &in) {
		return
	}
	// Serialize the entire read/modify/write so saving another connection
	// cannot restore a credential that was just disconnected.
	a.mu.Lock()
	defer a.mu.Unlock()
	previous := a.p
	p := previous
	if in.TextProvider != "" {
		if in.TextProvider != "openai" && in.TextProvider != "deepseek" && in.TextProvider != "aiservice" {
			apphost.Problem(w, 400, "Choose OpenAI, DeepSeek or AI Service")
			return
		}
		p.TextProvider = in.TextProvider
	}
	if in.ResearchProvider != "" {
		if in.ResearchProvider != "auto" && in.ResearchProvider != "openai" && in.ResearchProvider != "deepseek" && in.ResearchProvider != "feeds" {
			apphost.Problem(w, 400, "Choose a supported research source")
			return
		}
		p.ResearchProvider = in.ResearchProvider
	}
	if in.DeepSeekModel != "" {
		if !safeModel.MatchString(in.DeepSeekModel) {
			apphost.Problem(w, 400, "Enter a valid DeepSeek model ID")
			return
		}
		p.DeepSeekModel = in.DeepSeekModel
	}
	if in.DeepSeekKey != "" {
		if len(in.DeepSeekKey) > 512 || strings.ContainsAny(in.DeepSeekKey, "\r\n") {
			apphost.Problem(w, 400, "Invalid DeepSeek key")
			return
		}
		p.DeepSeekKey = strings.TrimSpace(in.DeepSeekKey)
	}
	if in.ServiceURL != "" {
		raw := strings.TrimRight(strings.TrimSpace(in.ServiceURL), "/")
		if !validServiceURL(raw) {
			apphost.Problem(w, 400, "AI Service must use a loopback HTTP address with an explicit port")
			return
		}
		if p.ServiceURL != raw {
			p.ServiceToken = ""
		}
		p.ServiceURL = raw
	}
	if in.ServiceToken != "" {
		token := strings.TrimSpace(in.ServiceToken)
		if len(token) < 32 || len(token) > 200 || strings.ContainsAny(token, "\r\n") {
			apphost.Problem(w, 400, "Enter the AI Service app credential issued for podcasts")
			return
		}
		p.ServiceToken = token
	}
	if in.ClearDeepSeek {
		p.DeepSeekKey = ""
	}
	if in.ClearService {
		p.ServiceToken = ""
	}
	normalizeProviders(&p)
	if in.Key != "" {
		if len(in.Key) > 512 || strings.ContainsAny(in.Key, "\r\n") {
			apphost.Problem(w, 400, "Invalid API key")
			return
		}
		p.Key = strings.TrimSpace(in.Key)
	}
	if in.XToken != "" {
		if len(in.XToken) > 2048 || strings.ContainsAny(in.XToken, "\r\n") {
			apphost.Problem(w, 400, "Invalid X token")
			return
		}
		p.XToken = strings.TrimSpace(in.XToken)
	}
	if in.ClearKey {
		p.Key = ""
	}
	if in.ClearX {
		p.XToken = ""
	}
	if (in.Model != "" && !safeModel.MatchString(in.Model)) || (in.TTSModel != "" && !safeModel.MatchString(in.TTSModel)) {
		apphost.Problem(w, 400, "Enter valid model IDs")
		return
	}
	if in.Model != "" {
		p.Model = in.Model
	}
	if in.TTSModel != "" {
		p.TTSModel = in.TTSModel
	}
	if in.SpeechProvider != "" {
		if in.SpeechProvider != "myvoice" && in.SpeechProvider != "openai" {
			apphost.Problem(w, 400, "Choose My Voice or OpenAI speech")
			return
		}
		p.SpeechProvider = in.SpeechProvider
	}
	if in.VoiceURL != "" {
		address := strings.TrimRight(strings.TrimSpace(in.VoiceURL), "/")
		if !validVoiceURL(address) {
			apphost.Problem(w, 400, "My Voice needs a loopback IP address and explicit port")
			return
		}
		if p.VoiceURL != address {
			p.VoiceToken = ""
		}
		p.VoiceURL = address
	}
	if in.VoiceToken != "" {
		p.VoiceToken = strings.TrimSpace(in.VoiceToken)
		if !validVoiceToken(p.VoiceToken) {
			apphost.Problem(w, 400, "Use the credential issued in My Voice Settings")
			return
		}
	}
	if in.ClearVoice {
		p.VoiceToken = ""
	}
	raw, _ := json.Marshal(p)
	enc, e := a.crypt(raw, true)
	if e == nil {
		_, e = a.h.DB.Exec("INSERT INTO podcast_providers VALUES(1,?) ON CONFLICT(id) DO UPDATE SET secret=excluded.secret", enc)
	}
	if e != nil {
		a.h.DBError(w, e)
		return
	}
	a.p = p
	if previous.VoiceToken != p.VoiceToken || previous.VoiceURL != p.VoiceURL {
		a.voice = voiceStatus{}
	}
	if revoked(previous.Key, p.Key) || revoked(previous.DeepSeekKey, p.DeepSeekKey) ||
		revoked(previous.ServiceToken, p.ServiceToken) || revoked(previous.XToken, p.XToken) ||
		(previous.ServiceToken != "" && previous.ServiceURL != p.ServiceURL) ||
		revoked(previous.VoiceToken, p.VoiceToken) || previous.SpeechProvider != p.SpeechProvider {
		for _, cancel := range a.active {
			cancel()
		}
	}
	a.h.Audit(apphost.Current(r).User.ID, "providers-updated", "podcasts")
	apphost.JSON(w, 200, map[string]any{"saved": true, "provider": a.providerInfoLocked()})
	a.signal()
}
func revoked(previous, current string) bool {
	return previous != "" && previous != current
}
func (a *App) loadSettings(user string) (Settings, error) {
	s := Defaults()
	var raw string
	e := a.h.DB.QueryRow("SELECT body FROM podcast_settings WHERE user_id=?", user).Scan(&raw)
	if e == sql.ErrNoRows {
		return s, nil
	}
	if e != nil {
		return s, e
	}
	e = json.Unmarshal([]byte(raw), &s)
	return s, e
}
func (a *App) settings(w http.ResponseWriter, r *http.Request) {
	var s Settings
	if !apphost.Decode(w, r, &s) {
		return
	}
	if e := validSettings(s); e != nil {
		apphost.Problem(w, 400, e.Error())
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	user := apphost.Current(r).User.ID
	var saved Settings
	var previous string
	var due int64
	err := a.h.DB.QueryRow("SELECT body,next_run FROM podcast_settings WHERE user_id=?", user).Scan(&previous, &due)
	if err == nil {
		err = json.Unmarshal([]byte(previous), &saved)
	}
	if err != nil && err != sql.ErrNoRows {
		a.h.DBError(w, err)
		return
	}
	changedSchedule := saved.Schedule != s.Schedule || saved.At != s.At
	if s.Schedule == "off" {
		due = 0
	} else if changedSchedule || due <= 0 {
		due = nextRun(s, time.Now())
	}
	raw, _ := json.Marshal(s)
	_, e := a.h.DB.Exec("INSERT INTO podcast_settings VALUES(?,?,?) ON CONFLICT(user_id) DO UPDATE SET body=excluded.body,next_run=excluded.next_run", user, string(raw), due)
	if e != nil {
		a.h.DBError(w, e)
		return
	}
	apphost.JSON(w, 200, map[string]bool{"saved": true})
	a.signal()
}
func (a *App) state(w http.ResponseWriter, r *http.Request) {
	u := apphost.Current(r).User
	s, e := a.loadSettings(u.ID)
	if e != nil {
		a.h.DBError(w, e)
		return
	}
	rows, e := a.h.DB.Query("SELECT body FROM podcast_episodes WHERE user_id=? ORDER BY created DESC LIMIT 50", u.ID)
	if e != nil {
		a.h.DBError(w, e)
		return
	}
	defer rows.Close()
	list := []Episode{}
	for rows.Next() {
		var raw string
		if e = rows.Scan(&raw); e != nil {
			a.h.DBError(w, e)
			return
		}
		var x Episode
		if e = json.Unmarshal([]byte(raw), &x); e != nil {
			a.h.DBError(w, e)
			return
		}
		x.Research = ""
		x.Script.Segments = nil
		list = append(list, x)
	}
	if rows.Err() != nil {
		a.h.DBError(w, rows.Err())
		return
	}
	a.mu.Lock()
	provider := a.providerInfoLocked()
	paused := a.p.productionError() != nil || a.speechIssueLocked(a.p) != ""
	a.mu.Unlock()
	var next int64
	_ = a.h.DB.QueryRow("SELECT next_run FROM podcast_settings WHERE user_id=?", u.ID).Scan(&next)
	apphost.JSON(w, 200, map[string]any{"settings": s, "episodes": list, "people": People, "feeds": Feeds, "nextRun": next, "schedulePaused": next > 0 && paused, "provider": provider, "limits": map[string]int{"daily": 3, "queued": 3, "library": 50}})
}
func (a *App) save(e *Episode) error {
	e.Updated = time.Now().Unix()
	raw, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = a.h.DB.Exec("UPDATE podcast_episodes SET status=?,updated=?,body=? WHERE id=?", e.Status, e.Updated, string(raw), e.ID)
	return err
}
func (a *App) load(id, user string) (Episode, error) {
	var raw string
	var e Episode
	err := a.h.DB.QueryRow("SELECT body FROM podcast_episodes WHERE id=? AND user_id=?", id, user).Scan(&raw)
	if err != nil {
		return e, err
	}
	err = json.Unmarshal([]byte(raw), &e)
	return e, err
}
func (a *App) getEpisode(w http.ResponseWriter, r *http.Request) {
	e, err := a.load(r.PathValue("id"), apphost.Current(r).User.ID)
	if err != nil {
		apphost.Problem(w, 404, "Episode not found")
		return
	}
	apphost.JSON(w, 200, e)
}
func (a *App) create(user string, s Settings) (Episode, error) {
	tx, err := a.h.DB.Begin()
	if err != nil {
		return Episode{}, err
	}
	defer tx.Rollback()
	e, err := createEpisode(tx, user, s, time.Now(), 0)
	if err != nil {
		return e, err
	}
	return e, tx.Commit()
}

var errAccountLimit = errors.New("Studio limit reached: up to 3 productions per 24 hours, 3 queued and 50 saved episodes")
var errQueueFull = errors.New("The studio queue is full. Please try again later")

func createEpisode(tx *sql.Tx, user string, s Settings, now time.Time, scheduledFor int64) (Episode, error) {
	e := Episode{ID: apphost.ID(), Created: now.Unix(), Updated: now.Unix(), ScheduledFor: scheduledFor, Status: "queued", Stage: "Waiting for the studio", Settings: s, Sources: []Source{}, Warnings: []string{}}
	var err error
	var pending, total, daily int
	if err = tx.QueryRow("SELECT count(*),sum(CASE WHEN status IN ('queued','running') THEN 1 ELSE 0 END) FROM podcast_episodes WHERE user_id=?", user).Scan(&total, nullableInt(&pending)); err != nil {
		return e, err
	}
	if err = tx.QueryRow("SELECT count(*) FROM podcast_productions WHERE user_id=? AND started>?", user, now.Add(-24*time.Hour).Unix()).Scan(&daily); err != nil {
		return e, err
	}
	if total >= 50 || pending >= 3 || daily >= 3 {
		return e, errAccountLimit
	}
	var global int
	if err = tx.QueryRow("SELECT count(*) FROM podcast_episodes WHERE status IN ('queued','running')").Scan(&global); err != nil {
		return e, err
	}
	if global >= 24 {
		return e, errQueueFull
	}
	raw, _ := json.Marshal(e)
	if _, err = tx.Exec("INSERT INTO podcast_episodes VALUES(?,?,?,?,?,?)", e.ID, user, e.Created, e.Updated, e.Status, string(raw)); err != nil {
		return e, err
	}
	if _, err = tx.Exec("INSERT INTO podcast_productions VALUES(?,?)", user, e.Created); err != nil {
		return e, err
	}
	if _, err = tx.Exec("DELETE FROM podcast_productions WHERE started<?", now.Add(-48*time.Hour).Unix()); err != nil {
		return e, err
	}
	return e, nil
}

type intScanner struct{ p *int }

func nullableInt(p *int) intScanner { return intScanner{p} }
func (s intScanner) Scan(v any) error {
	if v == nil {
		*s.p = 0
		return nil
	}
	if n, ok := v.(int64); ok {
		*s.p = int(n)
		return nil
	}
	return errors.New("invalid count")
}
func (a *App) enqueue(w http.ResponseWriter, r *http.Request) {
	if err := a.provider().productionError(); err != nil {
		apphost.Problem(w, 409, err.Error())
		return
	}
	s, err := a.loadSettings(apphost.Current(r).User.ID)
	if err != nil {
		a.h.DBError(w, err)
		return
	}
	a.mu.Lock()
	e, err := a.create(apphost.Current(r).User.ID, s)
	a.mu.Unlock()
	if err != nil {
		apphost.Problem(w, 409, err.Error())
		return
	}
	apphost.JSON(w, 202, e)
	a.signal()
}
func (a *App) cancel(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e, err := a.load(r.PathValue("id"), apphost.Current(r).User.ID)
	if err != nil {
		apphost.Problem(w, 404, "Episode not found")
		return
	}
	if e.Status != "running" && e.Status != "queued" {
		apphost.Problem(w, 409, "This episode is no longer queued or running")
		return
	}
	if cancel := a.active[e.ID]; cancel != nil {
		cancel()
	}
	e.Status = "cancelled"
	e.Stage = "Production cancelled"
	if err = a.save(&e); err != nil {
		a.h.DBError(w, err)
		return
	}
	apphost.JSON(w, 200, e)
}
func (a *App) retry(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e, err := a.load(r.PathValue("id"), apphost.Current(r).User.ID)
	if err != nil {
		apphost.Problem(w, 404, "Episode not found")
		return
	}
	if _, active := a.active[e.ID]; active {
		apphost.Problem(w, 409, "Production is still stopping; retry in a moment")
		return
	}
	if e.Status != "failed" && e.Status != "cancelled" && e.Status != "needs_review" && e.Status != "awaiting_voice" {
		apphost.Problem(w, 409, "Only interrupted or failed productions can resume")
		return
	}
	if !textComplete(e) {
		if err := a.p.productionError(); err != nil {
			apphost.Problem(w, 409, err.Error())
			return
		}
	}
	if textComplete(e) {
		if issue := a.speechIssueLocked(a.p); issue != "" {
			apphost.Problem(w, 409, issue)
			return
		}
	}
	var count int
	_ = a.h.DB.QueryRow("SELECT count(*) FROM podcast_episodes WHERE user_id=? AND status IN ('queued','running')", apphost.Current(r).User.ID).Scan(&count)
	if count >= 3 {
		apphost.Problem(w, 409, "Your production queue is full")
		return
	}
	var daily, global int
	if err = a.h.DB.QueryRow("SELECT count(*) FROM podcast_productions WHERE user_id=? AND started>?", apphost.Current(r).User.ID, time.Now().Add(-24*time.Hour).Unix()).Scan(&daily); err != nil {
		a.h.DBError(w, err)
		return
	}
	if err = a.h.DB.QueryRow("SELECT count(*) FROM podcast_episodes WHERE status IN ('queued','running')").Scan(&global); err != nil {
		a.h.DBError(w, err)
		return
	}
	if (daily >= 3 && e.Status != "awaiting_voice") || global >= 24 {
		apphost.Problem(w, 409, "Production limit reached; retry after your 24-hour allowance resets")
		return
	}
	if e.Status != "awaiting_voice" {
		if _, err = a.h.DB.Exec("INSERT INTO podcast_productions VALUES(?,?)", apphost.Current(r).User.ID, time.Now().Unix()); err != nil {
			a.h.DBError(w, err)
			return
		}
	}
	if e.Status == "needs_review" {
		e.Script = Script{}
		e.LengthNote = ""
		e.Review = Review{}
	}
	e.Status = "queued"
	e.Stage = "Resuming saved progress"
	e.Error = ""
	e.VoiceRetryAt = 0
	if err = a.save(&e); err != nil {
		a.h.DBError(w, err)
		return
	}
	a.signal()
	apphost.JSON(w, 202, e)
}
func (a *App) deleteEpisode(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	e, err := a.load(r.PathValue("id"), apphost.Current(r).User.ID)
	if err != nil {
		apphost.Problem(w, 404, "Episode not found")
		return
	}
	if e.Status == "running" || e.Status == "queued" || a.active[e.ID] != nil {
		apphost.Problem(w, 409, "Cancel production and wait for it to stop before deleting")
		return
	}
	dir := a.objectDir(e.ID)
	if err = os.RemoveAll(dir); err != nil {
		apphost.Problem(w, 500, "Episode audio could not be removed")
		return
	}
	if _, err = a.h.DB.Exec("DELETE FROM podcast_episodes WHERE id=?", e.ID); err != nil {
		a.h.DBError(w, err)
		return
	}
	apphost.JSON(w, 200, map[string]bool{"deleted": true})
}
func (a *App) objectDir(id string) string { return filepath.Join(a.h.Config.Data, "objects", id) }
func (a *App) audio(w http.ResponseWriter, r *http.Request) {
	e, err := a.load(r.PathValue("id"), apphost.Current(r).User.ID)
	if err != nil || e.Status != "ready" {
		apphost.Problem(w, 404, "Episode audio is not ready")
		return
	}
	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Cache-Control", "private, no-store")
	if r.URL.Query().Get("download") == "1" {
		w.Header().Set("Content-Disposition", "attachment; filename=\"podcast-"+e.ID[:12]+".wav\"")
	}
	http.ServeFile(w, r, filepath.Join(a.objectDir(e.ID), "episode.wav"))
}
func (a *App) transcript(w http.ResponseWriter, r *http.Request) {
	e, err := a.load(r.PathValue("id"), apphost.Current(r).User.ID)
	if err != nil {
		apphost.Problem(w, 404, "Episode not found")
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"transcript.md\"")
	io.WriteString(w, "# "+e.Script.Title+"\n\nAI-generated narrators. Research cutoff: "+e.Cutoff+"\n\n"+e.Script.Summary+"\n\n")
	for _, s := range e.Script.Segments {
		io.WriteString(w, "## "+s.Title+"\n\n")
		for _, t := range s.Turns {
			io.WriteString(w, "**"+t.Speaker+":** "+t.Text+"\n\nSources: "+strings.Join(t.Sources, ", ")+"\n\n")
		}
	}
	io.WriteString(w, "## Sources\n\n")
	for _, s := range e.Sources {
		io.WriteString(w, "- "+s.ID+": "+s.Title+" — "+s.URL+"\n")
	}
}
func (a *App) sourceExport(w http.ResponseWriter, r *http.Request) {
	e, err := a.load(r.PathValue("id"), apphost.Current(r).User.ID)
	if err != nil {
		apphost.Problem(w, 404, "Episode not found")
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\"sources.json\"")
	apphost.JSON(w, 200, map[string]any{"cutoff": e.Cutoff, "sources": e.Sources, "warnings": e.Warnings, "review": e.Review})
}
func (a *App) signal() {
	select {
	case a.wake <- struct{}{}:
	default:
	}
}
func validID(s string) bool { b, e := hex.DecodeString(s); return e == nil && len(b) == 24 }
