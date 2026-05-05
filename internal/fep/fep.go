package fep

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"

	"tea.kareha.org/cup/termi"
)

const bufferSize = 1024

type Engine interface {
	Init() error
	Finish() error
	Process(key termi.Key) (string, bool)
	Status() (string, bool)
}

type FEP struct {
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

func Init(cfg *Config, c *exec.Cmd, en Engine) (*FEP, error) {
	fgColor, err := termi.ParseHexColor(cfg.FgColor)
	if err != nil {
		return nil, err
	}
	bgColor, err := termi.ParseHexColor(cfg.BgColor)
	if err != nil {
		return nil, err
	}

	f, err := pty.Start(c)
	if err != nil {
		return nil, err
	}

	fep := &FEP{
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

	err = en.Init()
	if err != nil {
		reset()
		return nil, err
	}
	fep.drawStatus()

	go func() {
		for {
			key := termi.ReadKey()
			processed, update := fep.en.Process(key)
			if processed != "" {
				err = writeStringAll(f, processed)
				if err != nil {
					return
				}
			}
			if update {
				fep.drawStatus()
			}
		}
	}()

	go func() {
		c.Wait()
	}()

	listener := func(esc bool) {
		fep.esc = esc
		fep.drawStatus()
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
	err := fep.en.Finish()
	termi.RemoveEscapeListener(fep.listener)
	reset()
	return err
}

func (fep *FEP) drawStatus() {
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
