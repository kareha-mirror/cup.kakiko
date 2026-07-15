package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"tea.kareha.org/cup/termi/lock"
	"tea.kareha.org/cup/termi/shutil"

	"tea.kareha.org/cup/kakiko/internal/fep"
	"tea.kareha.org/cup/kakiko/internal/skk"
)

const appName = "kakiko"

var (
	//go:embed skk-edic-joyo.txt
	skkdicJoyo string

	//go:embed skk-edic-overlay.txt
	skkdicOverlay string
)

func fatal(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func getConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		fatal(err)
	}
	return filepath.Join(dir, appName)
}

func main() {
	configDir := flag.String("d", "", "config directory")
	joyo := flag.Bool("joyo", false, "joyo mode")
	unlock := flag.Bool("unlock", false, "unlock")
	download := flag.Bool("download", false, "download dictionary")
	flag.Parse()

	if *configDir == "" {
		*configDir = getConfigDir()
	}

	if *unlock {
		err := lock.Unlock(*configDir)
		if err != nil {
			fatal(err)
		}
		return
	}

	if *download {
		err := downloadDictionary(*configDir)
		if err != nil {
			fatal(err)
		}
		return
	}

	// duplication guard
	if os.Getenv("KAKIKO_RUNNING") != "" {
		fatal("already running")
	}
	os.Setenv("KAKIKO_RUNNING", "1")

	args := flag.Args()

	var command string
	var arguments []string
	if len(args) < 1 {
		command = shutil.Path()
	} else {
		command = args[0]
	}
	if len(args) > 1 {
		arguments = args[1:]
	}

	dics := []string{}
	if !*joyo {
		dics = append(dics, skkdicOverlay)
	}
	dics = append(dics, skkdicJoyo)
	en := skk.NewEngine(dics)

	f, err := fep.Init(*configDir, en, command, arguments...)
	if err != nil {
		fatal(err)
	}
	defer func() {
		err := f.Finish()
		if err != nil {
			fatal(err)
		}
	}()

	f.Main()
}
