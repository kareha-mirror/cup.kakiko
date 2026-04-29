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

const appName = "kakiko"

func dicPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, appName, "SKK-JISYO.L.cdb")
	return path, nil
}

func userDicPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, appName, "skk-jisyo")
	return path, nil
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

	path, err := dicPath()
	if err != nil {
		fatal(err)
	}
	userPath, err := userDicPath()
	if err != nil {
		fatal(err)
	}
	en := skk.NewEngine(path, userPath)
	f := fep.Init(c, en)
	defer f.Finish()
	f.Main()
}
