package podcasts

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"net/http"

	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
)

const maxIntroPCM = 60 * sampleRate * 2

const introSchema = `CREATE TABLE podcast_intro_tracks(
	user_id TEXT NOT NULL, digest TEXT NOT NULL, slot INTEGER, pcm BLOB NOT NULL,
	PRIMARY KEY(user_id,digest), UNIQUE(user_id,slot));
CREATE TABLE podcast_episode_intros(
	episode_id TEXT PRIMARY KEY REFERENCES podcast_episodes(id) ON DELETE CASCADE,
	user_id TEXT NOT NULL, digest TEXT NOT NULL,
	FOREIGN KEY(user_id,digest) REFERENCES podcast_intro_tracks(user_id,digest));
CREATE INDEX podcast_intro_usage ON podcast_episode_intros(user_id,digest);`

func readIntroPCM(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxIntroPCM+45))
	n := len(data) - 44
	if err != nil || n < sampleRate*2 || n > maxIntroPCM || n%2 != 0 ||
		!bytes.Equal(data[:44], wavHeader(uint32(n))) {
		return nil, errors.New("Choose an audio track between 1 and 60 seconds")
	}
	return data[44:], nil
}

func (a *App) setIntros(w http.ResponseWriter, r *http.Request) {
	user := apphost.Current(r).User.ID
	var tracks [2][]byte
	if r.Method == http.MethodPut {
		r.Body = http.MaxBytesReader(w, r.Body, 2*(maxIntroPCM+44)+8192)
		reader, err := r.MultipartReader()
		if err != nil {
			apphost.Problem(w, 400, "Choose two audio intro tracks")
			return
		}
		for i, field := range []string{"first", "second"} {
			part, err := reader.NextPart()
			if err != nil || part.FormName() != field {
				apphost.Problem(w, 400, "Choose two audio intro tracks")
				return
			}
			tracks[i], err = readIntroPCM(part)
			if err != nil {
				apphost.Problem(w, 400, err.Error())
				return
			}
		}
		if _, err := reader.NextPart(); err != io.EOF || bytes.Equal(tracks[0], tracks[1]) {
			apphost.Problem(w, 400, "Choose two different audio intro tracks")
			return
		}
	}
	tx, err := a.h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		a.h.DBError(w, err)
		return
	}
	defer tx.Rollback()
	if _, err = tx.Exec("UPDATE podcast_intro_tracks SET slot=NULL WHERE user_id=?", user); err != nil {
		a.h.DBError(w, err)
		return
	}
	for i, pcm := range tracks {
		if pcm == nil {
			break
		}
		hash := sha256.Sum256(pcm)
		digest := hex.EncodeToString(hash[:])
		if _, err = tx.Exec(`INSERT INTO podcast_intro_tracks VALUES(?,?,?,?)
			ON CONFLICT(user_id,digest) DO UPDATE SET slot=excluded.slot`, user, digest, i+1, pcm); err != nil {
			a.h.DBError(w, err)
			return
		}
	}
	if err = pruneIntroTracks(tx, user); err == nil {
		err = tx.Commit()
	}
	if err != nil {
		a.h.DBError(w, err)
		return
	}
	apphost.JSON(w, 200, map[string]bool{"introsEnabled": tracks[0] != nil})
}

func pruneIntroTracks(tx *sql.Tx, user string) error {
	_, err := tx.Exec(`DELETE FROM podcast_intro_tracks AS t WHERE user_id=? AND slot IS NULL
		AND NOT EXISTS(SELECT 1 FROM podcast_episode_intros e WHERE e.user_id=t.user_id AND e.digest=t.digest)`, user)
	return err
}
