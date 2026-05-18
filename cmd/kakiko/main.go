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
	"tea.kareha.org/cup/kakiko/internal/skkdic"
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

	var cfg *fep.Config
	cfgPath := getConfigPath(*configDir)
	_, err := os.Stat(cfgPath)
	if err != nil {
		cfg = fep.DefaultConfig()
		fep.SaveConfig(cfgPath, cfg)
	} else {
		cfg = fep.LoadConfig(cfgPath)
	}

	en := skk.NewEngine()

	if !*joyo {
		overlayDic := skkdic.NewStrDic(skkdicOverlay)
		en.AddDic(overlayDic)
	}

	joyoDic := skkdic.NewStrDic(skkdicJoyo)
	en.AddDic(joyoDic)

	dicPath := getSKKDicPath(*configDir)
	mainDic := skkdic.NewCDBDic(dicPath)
	en.AddDic(mainDic)

	userDicPath := getSKKUserDicPath(*configDir)
	userDic := skkdic.NewMemDic(userDicPath)
	en.SetUserDic(userDic)

	diffDicPath := getSKKDiffDicPath(*configDir)
	diffDic := skkdic.NewMemDic(diffDicPath)
	en.SetDiffDic(diffDic)

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
