package main

import (
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

func getConfigPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		fatal(err)
	}
	return filepath.Join(dir, appName, appName+".yaml")
}

func getSKKDicPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		fatal(err)
	}
	return filepath.Join(dir, appName, "skk-edic-legacy-l.cdb")
}

func getSKKUserDicPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		fatal(err)
	}
	return filepath.Join(dir, appName, "skk-edic-user.txt")
}

func getSKKDiffDicPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		fatal(err)
	}
	return filepath.Join(dir, appName, "skk-edic-diff.txt")
}

func main() {
	var command string
	var arguments []string
	if len(os.Args) < 2 {
		command = os.Getenv("SHELL")
		if command == "" {
			command = fallbackCommand
		}
	} else {
		command = os.Args[1]
	}
	if len(os.Args) > 2 {
		arguments = os.Args[2:]
	}
	var c = exec.Command(command, arguments...)

	var cfg *fep.Config
	cfgPath := getConfigPath()
	_, err := os.Stat(cfgPath)
	if err != nil {
		cfg = fep.DefaultConfig()
		fep.SaveConfig(cfgPath, cfg)
	} else {
		cfg = fep.LoadConfig(cfgPath)
	}

	dicPath := getSKKDicPath()
	userDicPath := getSKKUserDicPath()
	diffDicPath := getSKKDiffDicPath()
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
