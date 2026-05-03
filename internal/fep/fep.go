package fep

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"

	"tea.kareha.org/cup/termi"
)

type Engine interface {
	Process(key termi.Key) (string, bool)
	Status() (string, bool)
}

const bufferSize = 1024

type FEP struct {
	fd       *os.File
	en       Engine
	listener termi.EscapeListener
	esc      bool
}

func (f *FEP) updateSize() {
	rows, cols, err := pty.Getsize(os.Stdin)
	if err != nil {
		panic(err)
	}
	pty.Setsize(f.fd, &pty.Winsize{
		Rows: uint16(rows - 1),
		Cols: uint16(cols),
	})
}

func writeStringAll(fd *os.File, s string) error {
	data := []byte(s)
	total := 0

	for total < len(data) {
		n, err := fd.Write(data[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}

func Init(c *exec.Cmd, en Engine) *FEP {
	fd, err := pty.Start(c)
	if err != nil {
		panic(err)
	}

	f := &FEP{
		fd:       fd,
		en:       en,
		listener: nil,
		esc:      false,
	}

	f.updateSize()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			f.updateSize()
		}
	}()

	_, h := termi.Size()
	termi.ScrollRange(0, h-1)

	termi.Clear()
	termi.HomeCursor()
	termi.Raw()

	f.drawStatus()

	go func() {
		for {
			key := termi.ReadKey()
			processed, update := f.en.Process(key)
			if processed != "" {
				err = writeStringAll(fd, processed)
				if err != nil {
					return
				}
			}
			if update {
				f.drawStatus()
			}
		}
	}()

	go func() {
		c.Wait()
		os.Exit(0)
	}()

	listener := func(esc bool) {
		f.esc = esc
		f.drawStatus()
	}
	f.listener = termi.EscapeListener(&listener)
	termi.AddEscapeListener(f.listener)

	return f
}

func (f *FEP) Finish() {
	termi.RemoveEscapeListener(f.listener)

	termi.ScrollReset()

	termi.Clear()
	termi.HomeCursor()
	termi.Cooked()
	termi.ShowCursor()
}

func (f *FEP) drawStatus() {
	w, h := termi.Size()
	termi.SaveCursor()
	termi.HideCursor()
	termi.MoveCursor(0, h-1)

	termi.DefaultColor()

	status, inv := f.en.Status()
	if inv {
		termi.EnableInvert()
	}
	termi.Print(status)
	termi.ClearTail()
	if inv {
		termi.DisableInvert()
	}

	termi.MoveCursor(w-2, h-1)
	if f.esc {
		termi.Print(" *")
	} else {
		termi.Print(" .")
	}

	termi.ResetColor()

	termi.ShowCursor()
	termi.LoadCursor()
}

func (f *FEP) Main() {
	buf := make([]byte, bufferSize)
	for {
		n, err := f.fd.Read(buf)
		if err != nil {
			return
		}

		os.Stdout.Write(buf[:n])
	}
}
