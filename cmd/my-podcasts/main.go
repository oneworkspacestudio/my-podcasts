package main

import (
	"github.com/oneworkspacestudio/my-podcasts/internal/podcasts"
	"github.com/oneworkspacestudio/my-podcasts/internal/webui"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"os"
	"path/filepath"
	"strings"
)

var version = "0.1.0-preview.3"

func main() {
	has := false
	for _, v := range os.Args[1:] {
		if v == "--data" || strings.HasPrefix(v, "--data=") {
			has = true
		}
	}
	if !has {
		d, e := os.UserConfigDir()
		if e != nil {
			apphost.Report(e)
			os.Exit(1)
		}
		os.Args = append(os.Args, "--data", filepath.Join(d, "OneWorkspace", "MyPodcasts"))
	}
	if e := apphost.Main(apphost.Config{ID: "podcasts", Name: "My Podcasts", Version: version, Listen: "127.0.0.1:9044", RestoreFiles: []string{"providers.key"}}, webui.Files(), func(h *apphost.Host) (apphost.Application, error) { return podcasts.New(h) }); e != nil {
		apphost.Report(e)
		os.Exit(1)
	}
}
