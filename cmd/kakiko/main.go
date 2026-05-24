package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"tea.kareha.org/cup/kakiko/internal/fep"
	"tea.kareha.org/cup/kakiko/internal/skk"
)

const appName = "kakiko"
const fallbackCommand = "/bin/sh"

//go:embed skk-edic-joyo.txt
var skkdicJoyo string

//go:embed skk-edic-overlay.txt
var skkdicOverlay string

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
	flag.Parse()

	if *configDir == "" {
		*configDir = getConfigDir()
	}

	args := flag.Args()

	var command string
	var arguments []string
	if len(args) < 1 {
		command = os.Getenv("SHELL")
		if command == "" {
			command = fallbackCommand
		}
	} else {
		command = args[0]
	}
	if len(args) > 1 {
		arguments = args[1:]
	}
	var c = exec.Command(command, arguments...)

	dics := []string{}
	if !*joyo {
		dics = append(dics, skkdicOverlay)
	}
	dics = append(dics, skkdicJoyo)
	en := skk.NewEngine(dics)

	f, err := fep.Init(*configDir, en, c)
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
