package fep

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"

	"tea.kareha.org/cup/termi"
)

type Process func(key termi.Key) (string, bool)
type Status func() string

const defaultCommand = "/bin/sh"
const bufferSize = 1024

type FEP struct {
	fd       *os.File
	process  Process
	status   Status
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

func Init(args []string, process Process, status Status) *FEP {
	var command string = defaultCommand
	var arguments []string
	if len(args) > 1 {
		command = args[1]
	}
	if len(args) > 2 {
		arguments = args[2:]
	}
	var c = exec.Command(command, arguments...)
	fd, err := pty.Start(c)
	if err != nil {
		panic(err)
	}

	f := &FEP{
		fd:       fd,
		process:  process,
		status:   status,
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
			processed, update := f.process(key)
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

	status := f.status()
	termi.Print(status)
	termi.ClearTail()

	termi.MoveCursor(w-2, h-1)
	if f.esc {
		termi.Print(" *")
	} else {
		termi.Print(" .")
	}

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
