package podcasts

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
)

const sampleRate = 24000

func chunkName(p providers, s Settings, t Turn) string {
	voice := s.VoiceA
	if t.Speaker == "Jamie" {
		voice = s.VoiceB
	}
	h := sha256.Sum256([]byte(p.TTSModel + "\x00" + voice + "\x00" + speechInstructions(t.Speaker) + "\x00" + t.Text))
	return hex.EncodeToString(h[:]) + ".pcm"
}
func (a *App) narrate(ctx context.Context, p providers, e *Episode, t Turn) error {
	name := chunkName(p, e.Settings, t)
	if stat, err := os.Stat(filepath.Join(a.objectDir(e.ID), name)); err == nil && stat.Size() > 1000 && stat.Size()%2 == 0 {
		return nil
	}
	voice := e.Settings.VoiceA
	if t.Speaker == "Jamie" {
		voice = e.Settings.VoiceB
	}
	b, err := a.post(ctx, p, "/audio/speech", map[string]any{"model": p.TTSModel, "voice": voice, "input": cleanTranscript(t.Text), "instructions": speechInstructions(t.Speaker), "response_format": "pcm", "speed": 1.0}, 16<<20)
	if err != nil {
		return err
	}
	if len(b) < 1000 || len(b)%2 != 0 {
		return errors.New("The voice provider did not return valid PCM audio")
	}
	return a.writeObject(ctx, e.ID, name, b)
}

// Level each voice conservatively to -19 dBFS RMS with a -1 dBFS peak ceiling.
// This is deterministic amplitude balancing, not a claim of certified LUFS mastering.
func balance(b []byte) {
	n := len(b) / 2
	if n == 0 {
		return
	}
	sum := 0.0
	peak := 0.0
	for i := 0; i < n; i++ {
		v := float64(int16(binary.LittleEndian.Uint16(b[i*2:])))
		sum += v * v
		peak = math.Max(peak, math.Abs(v))
	}
	if peak == 0 {
		return
	}
	rms := math.Sqrt(sum / float64(n))
	gain := math.Min(32767*math.Pow(10, -19.0/20)/math.Max(rms, 1), math.Min(3.0, 29203/peak))
	fade := 120
	for i := 0; i < n; i++ {
		v := float64(int16(binary.LittleEndian.Uint16(b[i*2:]))) * gain
		if i < fade {
			v *= float64(i) / float64(fade)
		}
		if n-i < fade {
			v *= float64(n-i) / float64(fade)
		}
		binary.LittleEndian.PutUint16(b[i*2:], uint16(int16(math.Round(v))))
	}
}
func wavHeader(size uint32) []byte {
	h := make([]byte, 44)
	copy(h, "RIFF")
	binary.LittleEndian.PutUint32(h[4:], size+36)
	copy(h[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(h[16:], 16)
	binary.LittleEndian.PutUint16(h[20:], 1)
	binary.LittleEndian.PutUint16(h[22:], 1)
	binary.LittleEndian.PutUint32(h[24:], sampleRate)
	binary.LittleEndian.PutUint32(h[28:], sampleRate*2)
	binary.LittleEndian.PutUint16(h[32:], 2)
	binary.LittleEndian.PutUint16(h[34:], 16)
	copy(h[36:], "data")
	binary.LittleEndian.PutUint32(h[40:], size)
	return h
}
func (a *App) master(ctx context.Context, p providers, e *Episode) error {
	dir := a.objectDir(e.ID)
	// Only immutable audio belongs in objects/, which apphost snapshots live.
	// Build on the same private NTFS volume outside that snapshot allowlist,
	// then atomically finalize under the backup freeze lock.
	scratch := filepath.Join(a.h.Config.Data, "temp")
	if err := os.MkdirAll(scratch, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(scratch, ".master-")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	defer f.Close()
	if _, err = f.Write(make([]byte, 44)); err != nil {
		return err
	}
	var size int64
	for i := range e.Script.Segments {
		e.Script.Segments[i].Start = float64(size) / float64(sampleRate*2)
		for _, t := range e.Script.Segments[i].Turns {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			b, err := os.ReadFile(filepath.Join(dir, chunkName(p, e.Settings, t)))
			if err != nil {
				return err
			}
			balance(b)
			if _, err = f.Write(b); err != nil {
				return err
			}
			size += int64(len(b))
			silence := make([]byte, sampleRate*2*17/100)
			if _, err = f.Write(silence); err != nil {
				return err
			}
			size += int64(len(silence))
		}
		silence := make([]byte, sampleRate)
		if _, err = f.Write(silence); err != nil {
			return err
		}
		size += int64(len(silence))
	}
	if size < sampleRate*2 || size > uint32Max-36 {
		return errors.New("Audio output exceeded the supported length")
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err = f.Write(wavHeader(uint32(size))); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	a.h.Freeze.RLock()
	defer a.h.Freeze.RUnlock()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	final := filepath.Join(dir, "episode.wav")
	if _, err = os.Stat(final); err == nil {
		if err = os.Remove(final); err != nil {
			return err
		}
	}
	if err = os.Rename(name, final); err != nil {
		return err
	}
	e.Duration = float64(size) / float64(sampleRate*2)
	e.AudioBytes = size + 44
	return nil
}

const uint32Max = int64(1<<32 - 1)
