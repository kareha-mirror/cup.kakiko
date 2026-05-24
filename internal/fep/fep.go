package fep

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/creack/pty"

	"tea.kareha.org/cup/termi"
)

const bufferSize = 1024

type Cmd int

const (
	CmdNone = iota
	CmdDraw
	CmdSync
)

type Engine interface {
	Init(dir string) error
	Finish() error
	Process(key termi.Key) (string, Cmd)
	Status() (string, bool)
	Sync() error
}

type FEP struct {
	dir     string
	cfg     *Config
	fgColor termi.Color
	bgColor termi.Color

	f        *os.File
	en       Engine
	listener termi.EscapeListener
	esc      bool
}

func (fep *FEP) updateSize() error {
	rows, cols, err := pty.Getsize(os.Stdin)
	if err != nil {
		return err
	}
	pty.Setsize(fep.f, &pty.Winsize{
		Rows: uint16(rows - 1),
		Cols: uint16(cols),
	})
	return nil
}

func writeStringAll(f *os.File, s string) error {
	data := []byte(s)
	total := 0

	for total < len(data) {
		n, err := f.Write(data[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}

func getConfigPath(dir string) string {
	return filepath.Join(dir, "fep.yaml")
}

func getLockPath(dir string) string {
	return filepath.Join(dir, "lock")
}

func lock(dir string) error {
	path := getLockPath(dir)
	for i := 0; i < 8; i++ {
		err := os.Mkdir(path, 0777)
		if err == nil {
			return nil
		}
		d, _ := time.ParseDuration("1s")
		time.Sleep(d)
	}
	return fmt.Errorf("cannot create lock")
}

func unlock(dir string) error {
	path := getLockPath(dir)
	for i := 0; i < 8; i++ {
		err := os.Remove(path)
		if err == nil {
			return nil
		}
		d, _ := time.ParseDuration("1s")
		time.Sleep(d)
	}
	return fmt.Errorf("cannot remove lock")
}

func Init(dir string, en Engine, c *exec.Cmd) (*FEP, error) {
	var cfg *Config
	cfgPath := getConfigPath(dir)
	_, err := os.Stat(cfgPath)
	if err != nil {
		cfg = DefaultConfig()
		SaveConfig(cfgPath, cfg)
	} else {
		cfg = LoadConfig(cfgPath)
	}

	fgColor, err := termi.ParseColor(cfg.FgColor)
	if err != nil {
		return nil, err
	}

	bgColor, err := termi.ParseColor(cfg.BgColor)
	if err != nil {
		return nil, err
	}

	f, err := pty.Start(c)
	if err != nil {
		return nil, err
	}

	fep := &FEP{
		dir:     dir,
		cfg:     cfg,
		fgColor: fgColor,
		bgColor: bgColor,

		f:        f,
		en:       en,
		listener: nil,
		esc:      false,
	}

	err = fep.updateSize()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	if err == nil {
		go func() {
			for range ch {
				err := fep.updateSize()
				if err != nil {
					return
				}
			}
		}()
	}

	_, h := termi.Size()
	termi.ScrollRange(0, h-1)

	termi.Clear()
	termi.HomeCursor()
	termi.Raw()

	err = lock(dir)
	if err != nil {
		reset()
		return nil, err
	}
	err = en.Init(dir)
	unlock(dir)
	if err != nil {
		reset()
		return nil, err
	}

	fep.draw()

	go func() {
		for {
			key := termi.ReadKey()
			processed, cmd := en.Process(key)
			if processed != "" {
				err = writeStringAll(f, processed)
				if err != nil {
					return
				}
			}
			switch cmd {
			case CmdDraw:
				fep.draw()
			case CmdSync:
				fep.sync()
				fep.draw()
			}
		}
	}()

	go func() {
		c.Wait()
	}()

	listener := func(esc bool) {
		fep.esc = esc
		fep.draw()
	}
	fep.listener = termi.EscapeListener(&listener)
	termi.AddEscapeListener(fep.listener)

	return fep, nil
}

func reset() {
	termi.ScrollReset()
	termi.Clear()
	termi.HomeCursor()
	termi.Cooked()
	termi.ShowCursor()
}

func (fep *FEP) Finish() error {
	err := lock(fep.dir)
	if err == nil {
		err = fep.en.Finish()
		unlock(fep.dir)
	}

	termi.RemoveEscapeListener(fep.listener)
	reset()
	return err
}

func (fep *FEP) sync() error {
	err := lock(fep.dir)
	if err != nil {
		return err
	}
	err = fep.en.Sync()
	unlock(fep.dir)
	return err
}

func (fep *FEP) draw() {
	w, h := termi.Size()
	termi.SaveCursor()
	termi.HideCursor()
	termi.MoveCursor(0, h-1)

	//termi.DefaultColor()
	termi.SetFgColor(fep.fgColor)
	termi.SetBgColor(fep.bgColor)

	status, inv := fep.en.Status()
	if inv {
		termi.EnableInvert()
	}
	termi.Print(status)
	termi.ClearTail()
	if inv {
		termi.DisableInvert()
	}

	termi.MoveCursor(w-2, h-1)
	if fep.esc {
		termi.Print(" *")
	} else {
		termi.Print(" .")
	}

	termi.ResetColor()

	termi.ShowCursor()
	termi.LoadCursor()
}

func (fep *FEP) Main() {
	buf := make([]byte, bufferSize)
	for {
		n, err := fep.f.Read(buf)
		if err != nil {
			return
		}

		os.Stdout.Write(buf[:n])
	}
}
