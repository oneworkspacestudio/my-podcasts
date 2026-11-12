package podcasts

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func (a *App) Run(ctx context.Context) error {
	t := time.NewTicker(20 * time.Second)
	defer t.Stop()
	return a.run(ctx, time.Now, t.C)
}

func (a *App) run(ctx context.Context, now func() time.Time, ticks <-chan time.Time) error {
	// A single supervised worker resumes persisted stages after a service restart.
	rows, err := a.h.DB.Query("SELECT id,body FROM podcast_episodes WHERE status='running'")
	if err != nil {
		return err
	}
	recoveries := []Episode{}
	for rows.Next() {
		var id, raw string
		if err = rows.Scan(&id, &raw); err != nil {
			rows.Close()
			return err
		}
		var e Episode
		if err = json.Unmarshal([]byte(raw), &e); err != nil {
			rows.Close()
			return err
		}
		e.Status = "queued"
		e.Stage = "Resuming after a restart"
		recoveries = append(recoveries, e)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, e := range recoveries {
		if err = a.save(&e); err != nil {
			return err
		}
	}
	for {
		if ctx.Err() != nil {
			return nil
		}
		a.refreshVoice(ctx)
		if err = a.schedule(now()); err != nil {
			a.h.Log.Error("podcast scheduler unavailable")
		}
		found, err := a.work(ctx)
		if err != nil {
			return err
		}
		if found {
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticks:
		case <-a.wake:
		}
	}
}
func (a *App) schedule(now time.Time) error {
	a.h.Freeze.RLock()
	defer a.h.Freeze.RUnlock()
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.speechIssueLocked(a.p) != "" {
		return nil
	}
	if a.p.SpeechProvider == "myvoice" {
		if err := a.resumeLocalSpeech(now); err != nil {
			return err
		}
	}
	if a.p.productionError() != nil {
		return nil
	}
	rows, err := a.h.DB.Query("SELECT s.user_id,s.body,s.next_run FROM podcast_settings s JOIN host_users u ON u.id=s.user_id WHERE s.next_run>0 AND s.next_run<=? AND u.role!='disabled' ORDER BY s.next_run,s.user_id", now.Unix())
	if err != nil {
		return err
	}
	type due struct {
		user string
		s    Settings
		at   int64
	}
	list := []due{}
	for rows.Next() {
		var user, raw string
		var at int64
		if err = rows.Scan(&user, &raw, &at); err != nil {
			rows.Close()
			return err
		}
		var s Settings
		if err = json.Unmarshal([]byte(raw), &s); err != nil {
			rows.Close()
			return err
		}
		list = append(list, due{user, s, at})
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, d := range list {
		if validSettings(d.s) != nil {
			continue
		}
		if err := a.enqueueScheduled(d.user, d.s, now, d.at); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) resumeLocalSpeech(now time.Time) error {
	var raw string
	err := a.h.DB.QueryRow(`SELECT e.body FROM podcast_episodes e JOIN host_users u ON u.id=e.user_id
WHERE e.status='awaiting_voice' AND u.role!='disabled' AND coalesce(json_extract(e.body,'$.voiceRetryAt'),0)<=?
AND NOT EXISTS(SELECT 1 FROM podcast_episodes q WHERE q.user_id=e.user_id AND q.status IN ('queued','running'))
AND (SELECT count(*) FROM podcast_episodes WHERE status IN ('queued','running'))<24
ORDER BY e.created,e.id LIMIT 1`, now.Unix()).Scan(&raw)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	var e Episode
	if err := json.Unmarshal([]byte(raw), &e); err != nil {
		return err
	}
	e.Status, e.Stage, e.Error, e.VoiceRetryAt = "queued", "Resuming with My Voice", "", 0
	return a.save(&e)
}
func (a *App) enqueueScheduled(user string, settings Settings, now time.Time, scheduledFor int64) error {
	tx, err := a.h.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var pending bool
	if err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM podcast_episodes WHERE user_id=? AND status IN ('queued','running'))", user).Scan(&pending); err != nil {
		return err
	}
	if pending {
		return nil
	}
	next := nextRun(settings, now)
	if settings.Schedule == "daily-three" {
		// After downtime, catch up on the latest edition and skip older slots.
		scheduledFor = next - int64(threeDailySpacing/time.Second)
	}
	if _, err := createEpisode(tx, user, settings, now, scheduledFor); err != nil {
		if errors.Is(err, errAccountLimit) || errors.Is(err, errQueueFull) {
			return nil
		}
		return err
	}
	if _, err := tx.Exec("UPDATE podcast_settings SET next_run=? WHERE user_id=?", next, user); err != nil {
		return err
	}
	return tx.Commit()
}
func (a *App) checkpoint(ctx context.Context, e *Episode) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.h.Freeze.RLock()
	defer a.h.Freeze.RUnlock()
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	var status, role string
	if err := a.h.DB.QueryRow("SELECT e.status,u.role FROM podcast_episodes e JOIN host_users u ON u.id=e.user_id WHERE e.id=?", e.ID).Scan(&status, &role); err != nil {
		return err
	}
	if status == "cancelled" || role == "disabled" {
		return context.Canceled
	}
	return a.save(e)
}
func (a *App) work(root context.Context) (bool, error) {
	a.h.Freeze.RLock()
	a.mu.Lock()
	var raw, user string
	err := a.h.DB.QueryRow("SELECT e.body,e.user_id FROM podcast_episodes e JOIN host_users u ON u.id=e.user_id WHERE e.status='queued' AND u.role!='disabled' ORDER BY e.created,e.id LIMIT 1").Scan(&raw, &user)
	if err == sql.ErrNoRows {
		a.mu.Unlock()
		a.h.Freeze.RUnlock()
		return false, nil
	}
	if err != nil {
		a.mu.Unlock()
		a.h.Freeze.RUnlock()
		return false, err
	}
	var e Episode
	if err = json.Unmarshal([]byte(raw), &e); err != nil {
		a.mu.Unlock()
		a.h.Freeze.RUnlock()
		return false, err
	}
	ctx, cancel := context.WithTimeout(root, 55*time.Minute)
	p := a.p
	a.active[e.ID] = cancel
	e.Status = "running"
	err = a.save(&e)
	a.mu.Unlock()
	a.h.Freeze.RUnlock()
	if err != nil {
		cancel()
		return false, err
	}
	ctx = context.WithValue(ctx, generationIdentity{}, user)
	err = a.produce(ctx, p, &e)
	cancel()
	var saveErr error
	a.h.Freeze.RLock()
	a.mu.Lock()
	delete(a.active, e.ID)
	var status string
	queryErr := a.h.DB.QueryRow("SELECT status FROM podcast_episodes WHERE id=? AND user_id=?", e.ID, user).Scan(&status)
	if queryErr == nil && status != "cancelled" {
		if root.Err() != nil {
			e.Status = "queued"
			e.Stage = "Paused for service restart"
			e.Error = ""
		} else if err != nil {
			if errors.Is(err, errReview) {
				e.Status = "needs_review"
				e.Stage = "Editorial review needs attention"
			} else if errors.Is(err, context.Canceled) {
				e.Status = "cancelled"
				e.Stage = "Production stopped"
			} else {
				e.Status = "failed"
				e.Stage = "Production interrupted"
			}
			e.Error = err.Error()
		}
		saveErr = a.save(&e)
	}
	a.mu.Unlock()
	a.h.Freeze.RUnlock()
	if queryErr != nil {
		return true, queryErr
	}
	if saveErr != nil {
		return true, saveErr
	}
	if root.Err() != nil {
		return true, nil
	}
	return true, nil
}

var errReview = errors.New("The editorial check found issues. Review the transcript and research, then regenerate before narration")

func (a *App) produce(ctx context.Context, p providers, e *Episode) error {
	if !textComplete(*e) {
		if err := p.productionError(); err != nil {
			return err
		}
	}
	if !validID(e.ID) {
		return errors.New("Invalid episode ID")
	}
	if e.Research == "" {
		e.Stage = "Researching reporting and public posts"
		e.Progress = 8
		if err := a.checkpoint(ctx, e); err != nil {
			return err
		}
		if err := a.research(ctx, p, e); err != nil {
			return err
		}
		if err := a.checkpoint(ctx, e); err != nil {
			return err
		}
	}
	if len(e.Script.Segments) == 0 {
		e.Stage = "Writing the conversation and chapters"
		e.Progress = 27
		if err := a.checkpoint(ctx, e); err != nil {
			return err
		}
		s, note, err := a.writeScript(ctx, p, e)
		if err != nil {
			return err
		}
		e.WriterProvider = p.TextProvider
		e.Script = s
		e.LengthNote = note
		if err = a.checkpoint(ctx, e); err != nil {
			return err
		}
	}
	if !e.Review.Passed {
		e.Stage = "Checking claims, attribution and pacing"
		e.Progress = 40
		if err := a.checkpoint(ctx, e); err != nil {
			return err
		}
		review, err := a.editorial(ctx, p, e)
		if err != nil {
			return err
		}
		e.Review = review
		if err = a.checkpoint(ctx, e); err != nil {
			return err
		}
		if !review.Passed {
			return errReview
		}
	}
	if p.SpeechProvider == "myvoice" {
		status := a.probeVoice(ctx, p)
		if status.Model == "" {
			return a.awaitVoice(ctx, e, status.Issue)
		}
		p.TTSModel = status.Model
		p.voiceCast = status.Cast
	} else if p.Key == "" {
		return a.awaitVoice(ctx, e, "Connect voices in Connections. Your research and reviewed script are saved")
	}
	if err := os.MkdirAll(a.objectDir(e.ID), 0700); err != nil {
		return err
	}
	total := 0
	for _, s := range e.Script.Segments {
		total += len(s.Turns)
	}
	done := 0
	for _, s := range e.Script.Segments {
		for _, t := range s.Turns {
			e.Stage = fmt.Sprintf("Recording %s · take %d of %d", t.Speaker, done+1, total)
			e.Progress = 45 + done*45/total
			if err := a.checkpoint(ctx, e); err != nil {
				return err
			}
			if err := a.narrate(ctx, p, e, t); err != nil {
				if errors.Is(err, errVoiceUnavailable) {
					return a.awaitVoice(ctx, e, "My Voice is busy or unavailable. Recording will resume automatically")
				}
				return err
			}
			done++
		}
	}
	e.Stage = "Balancing voices and finishing your episode"
	e.Progress = 93
	if err := a.checkpoint(ctx, e); err != nil {
		return err
	}
	if err := a.master(ctx, p, e); err != nil {
		return err
	}
	e.Status = "ready"
	e.Stage = "Ready to listen"
	e.Progress = 100
	e.Error = ""
	e.VoiceRetryAt = 0
	return a.checkpoint(ctx, e)
}

func (a *App) awaitVoice(ctx context.Context, e *Episode, issue string) error {
	e.Status = "awaiting_voice"
	e.Stage = "Script ready · waiting for narration"
	e.Progress = max(e.Progress, 45)
	e.Error = issue
	e.VoiceRetryAt = time.Now().Add(time.Minute).Unix()
	return a.checkpoint(ctx, e)
}
func (a *App) writeObject(ctx context.Context, id, name string, b []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.h.Freeze.RLock()
	defer a.h.Freeze.RUnlock()
	dir := a.objectDir(id)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".writing-")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, name))
}
