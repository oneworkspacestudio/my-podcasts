package podcasts

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Person struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Company string `json:"company"`
	Handle  string `json:"handle"`
	URL     string `json:"url"`
}

var People = []Person{
	{"sama", "Sam Altman", "OpenAI", "sama", "https://x.com/sama"},
	{"satya", "Satya Nadella", "Microsoft", "satyanadella", "https://x.com/satyanadella"},
	{"sundar", "Sundar Pichai", "Google", "sundarpichai", "https://x.com/sundarpichai"},
	{"demis", "Demis Hassabis", "Google DeepMind", "demishassabis", "https://x.com/demishassabis"},
	{"elon", "Elon Musk", "Tesla · xAI", "elonmusk", "https://x.com/elonmusk"},
	{"andrej", "Andrej Karpathy", "Independent AI researcher", "karpathy", "https://x.com/karpathy"},
	{"yann", "Yann LeCun", "AI research", "ylecun", "https://x.com/ylecun"},
	{"mustafa", "Mustafa Suleyman", "Microsoft AI", "mustafasuleyman", "https://x.com/mustafasuleyman"},
}

type Feed struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Kind string `json:"kind"`
}

var Feeds = []Feed{
	{"bbc", "BBC Technology", "https://feeds.bbci.co.uk/news/technology/rss.xml", "reporting"},
	{"verge", "The Verge", "https://www.theverge.com/rss/index.xml", "reporting"},
	{"techcrunch", "TechCrunch", "https://techcrunch.com/feed/", "reporting"},
	{"ars", "Ars Technica", "https://feeds.arstechnica.com/arstechnica/technology-lab", "reporting"},
	{"google", "Google Blog", "https://blog.google/rss/", "primary"},
	{"microsoft", "Microsoft Blog", "https://blogs.microsoft.com/feed/", "primary"},
	{"nvidia", "NVIDIA Blog", "https://blogs.nvidia.com/feed/", "primary"},
}

type Settings struct {
	ShowName      string   `json:"showName"`
	Topics        string   `json:"topics"`
	People        []string `json:"people"`
	Feeds         []string `json:"feeds"`
	Format        string   `json:"format"`
	Minutes       int      `json:"minutes"`
	LookbackHours int      `json:"lookbackHours"`
	VoiceA        string   `json:"voiceA"`
	VoiceB        string   `json:"voiceB"`
	Schedule      string   `json:"schedule"`
	At            string   `json:"at"`
}

func Defaults() Settings {
	return Settings{"The Tech Signal", "AI, big tech, the people building the future, and what it means for everyday life", []string{"sama", "satya", "sundar", "demis", "andrej"}, []string{"bbc", "verge", "techcrunch", "google", "microsoft", "nvidia"}, "conversation", 15, 48, "marin", "cedar", "off", "07:00"}
}
func validSettings(s Settings) error {
	if len(strings.TrimSpace(s.ShowName)) < 1 || len(s.ShowName) > 90 || len(s.Topics) > 1200 {
		return errors.New("Use a show name up to 90 characters and a focus up to 1,200 characters")
	}
	if s.Format != "briefing" && s.Format != "conversation" && s.Format != "deepdive" {
		return errors.New("Choose a supported episode format")
	}
	if s.Minutes != 8 && s.Minutes != 15 && s.Minutes != 25 && s.Minutes != 40 {
		return errors.New("Choose 8, 15, 25 or 40 minutes")
	}
	if s.LookbackHours != 24 && s.LookbackHours != 48 && s.LookbackHours != 168 {
		return errors.New("Choose a 24-hour, 48-hour or seven-day research window")
	}
	voices := map[string]bool{"marin": true, "cedar": true, "coral": true, "alloy": true, "sage": true, "ash": true}
	if !voices[s.VoiceA] || !voices[s.VoiceB] {
		return errors.New("Choose supported narrator voices")
	}
	if s.VoiceA == s.VoiceB {
		return errors.New("Choose two distinct voices")
	}
	if s.Schedule != "off" && s.Schedule != "daily" && s.Schedule != "daily-three" && s.Schedule != "weekly" {
		return errors.New("Choose on demand, once daily, three times daily, or weekly scheduling")
	}
	if _, e := time.Parse("15:04", s.At); e != nil {
		return errors.New("Choose a valid UTC production time")
	}
	ids := map[string]bool{}
	for _, p := range People {
		ids[p.ID] = true
	}
	seen := map[string]bool{}
	for _, id := range s.People {
		if !ids[id] || seen[id] {
			return errors.New("Choose people from the source directory")
		}
		seen[id] = true
	}
	ids = map[string]bool{}
	seen = map[string]bool{}
	for _, f := range Feeds {
		ids[f.ID] = true
	}
	for _, id := range s.Feeds {
		if !ids[id] || seen[id] {
			return errors.New("Choose publications from the source directory")
		}
		seen[id] = true
	}
	if len(s.Feeds) == 0 {
		return errors.New("Choose at least one publication")
	}
	return nil
}

const threeDailySpacing = 8 * time.Hour

func nextRun(s Settings, now time.Time) int64 {
	if s.Schedule == "off" {
		return 0
	}
	t, _ := time.Parse("15:04", s.At)
	now = now.UTC()
	hour := t.Hour()
	if s.Schedule == "daily-three" {
		hour %= 8
	}
	n := time.Date(now.Year(), now.Month(), now.Day(), hour, t.Minute(), 0, 0, time.UTC)
	if s.Schedule == "daily-three" {
		if !n.After(now) {
			n = n.Add((now.Sub(n)/threeDailySpacing + 1) * threeDailySpacing)
		}
		return n.Unix()
	}
	if !n.After(now) {
		n = n.AddDate(0, 0, 1)
	}
	if s.Schedule == "weekly" {
		for n.Weekday() != time.Monday {
			n = n.AddDate(0, 0, 1)
		}
	}
	return n.Unix()
}

type Source struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Publisher string `json:"publisher"`
	Kind      string `json:"kind"`
	Published string `json:"published"`
	Excerpt   string `json:"excerpt,omitempty"`
}
type Turn struct {
	Speaker string   `json:"speaker"`
	Text    string   `json:"text"`
	Sources []string `json:"sources"`
}
type Segment struct {
	Title string  `json:"title"`
	Kind  string  `json:"kind"`
	Turns []Turn  `json:"turns"`
	Start float64 `json:"start"`
}
type Script struct {
	Title    string    `json:"title"`
	Summary  string    `json:"summary"`
	Segments []Segment `json:"segments"`
}
type Review struct {
	Passed bool     `json:"passed"`
	Issues []string `json:"issues"`
}
type Episode struct {
	WriterProvider   string   `json:"writerProvider,omitempty"`
	ResearchProvider string   `json:"researchProvider,omitempty"`
	ID               string   `json:"id"`
	Created          int64    `json:"created"`
	ScheduledFor     int64    `json:"scheduledFor,omitempty"`
	Updated          int64    `json:"updated"`
	Status           string   `json:"status"`
	Stage            string   `json:"stage"`
	Progress         int      `json:"progress"`
	Error            string   `json:"error"`
	Settings         Settings `json:"settings"`
	Cutoff           string   `json:"cutoff"`
	Sources          []Source `json:"sources"`
	Warnings         []string `json:"warnings"`
	Research         string   `json:"research,omitempty"`
	Script           Script   `json:"script"`
	LengthNote       string   `json:"lengthNote,omitempty"`
	Review           Review   `json:"review"`
	Duration         float64  `json:"duration"`
	AudioBytes       int64    `json:"audioBytes"`
	VoiceRetryAt     int64    `json:"voiceRetryAt,omitempty"`
}

var safeModel = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,79}$`)

func webURL(raw string) bool {
	u, e := url.Parse(raw)
	return e == nil && u.Scheme == "https" && u.User == nil && u.Host != "" && u.Port() == "" && !strings.ContainsAny(raw, "\r\n\\")
}

const minimumEpisodeWords = 145

var errScriptTooLong = errors.New("The script exceeds the maximum supported episode length")

func validateScript(s Script, sources []Source) (int, error) {
	if len(s.Title) < 3 || len(s.Title) > 180 || len(s.Summary) > 1800 || len(s.Segments) < 3 || len(s.Segments) > 14 {
		return 0, errors.New("The script did not have a complete episode structure")
	}
	ids := map[string]bool{}
	for _, x := range sources {
		ids[x.ID] = true
	}
	words := 0
	turns := 0
	counts := map[string]int{}
	speakers := map[string]bool{}
	if s.Segments[0].Kind != "hook" || s.Segments[len(s.Segments)-1].Kind != "outro" {
		return 0, errors.New("The episode needs a cold open and a clear ending")
	}
	for _, seg := range s.Segments {
		if seg.Kind != "hook" && seg.Kind != "intro" && seg.Kind != "story" && seg.Kind != "analysis" && seg.Kind != "outro" {
			return 0, errors.New("Unknown chapter type")
		}
		counts[seg.Kind]++
		if len(seg.Title) == 0 || len(seg.Title) > 160 || len(seg.Turns) == 0 || len(seg.Turns) > 48 {
			return 0, errors.New("Invalid episode chapter")
		}
		for _, t := range seg.Turns {
			speakers[t.Speaker] = true
			turns++
			words += len(strings.Fields(t.Text))
			if t.Speaker != "Alex" && t.Speaker != "Jamie" {
				return 0, errors.New("Unknown narrator")
			}
			if strings.TrimSpace(t.Text) == "" || len(t.Text) > 3500 {
				return 0, errors.New("Narration chunk was too long or empty")
			}
			if (seg.Kind == "story" || seg.Kind == "analysis") && len(t.Sources) == 0 {
				return 0, errors.New("A story turn was missing its evidence links")
			}
			for _, id := range t.Sources {
				if !ids[id] {
					return 0, errors.New("The script referenced evidence outside its research dossier")
				}
			}
		}
	}
	if counts["intro"] == 0 || counts["story"]+counts["analysis"] == 0 || len(speakers) != 2 {
		return 0, errors.New("The episode needs an introduction, reported story and two narrators")
	}
	if turns > 220 || words > 7400 {
		return words, errScriptTooLong
	}
	return words, nil
}
