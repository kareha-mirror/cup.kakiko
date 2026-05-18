package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"tea.kareha.org/cup/kakiko/internal/fep"
	"tea.kareha.org/cup/kakiko/internal/skk"
)

const fallbackCommand = "/bin/sh"

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

func getConfigPath(dir string) string {
	return filepath.Join(dir, appName+".yaml")
}

func getSKKDicPath(dir string) string {
	return filepath.Join(dir, "skk-edic-legacy-l.cdb")
}

func getSKKUserDicPath(dir string) string {
	return filepath.Join(dir, "skk-edic-user.txt")
}

func getSKKDiffDicPath(dir string) string {
	return filepath.Join(dir, "skk-edic-diff.txt")
}

func main() {
	configDir := flag.String("d", "", "config directory")
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

	var cfg *fep.Config
	cfgPath := getConfigPath(*configDir)
	_, err := os.Stat(cfgPath)
	if err != nil {
		cfg = fep.DefaultConfig()
		fep.SaveConfig(cfgPath, cfg)
	} else {
		cfg = fep.LoadConfig(cfgPath)
	}

	dicPath := getSKKDicPath(*configDir)
	userDicPath := getSKKUserDicPath(*configDir)
	diffDicPath := getSKKDiffDicPath(*configDir)
	en := skk.NewEngine(skkEdicDefault, dicPath, userDicPath, diffDicPath)

	f, err := fep.Init(cfg, en, c)
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
