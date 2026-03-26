package fep

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"

	"tea.kareha.org/lab/termi"
)

type Process func(b []byte) []byte
type Status func() string

const defaultCommand = "/bin/sh"
const bufferSize = 1024

type FEP struct {
	fd      *os.File
	process Process
	status  Status
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
		fd:      fd,
		process: process,
		status:  status,
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

	go func() {
		for {
			b := make([]byte, 1)
			n, err := os.Stdin.Read(b)
			if err != nil {
				return
			}
			if n == 0 {
				continue
			}

			processed := f.process(b)

			_, err = fd.Write(processed)
			if err != nil {
				return
			}
		}
	}()

	go func() {
		c.Wait()
		os.Exit(0)
	}()

	return f
}

func (f *FEP) Finish() {
	_, h := termi.Size()
	termi.ScrollRange(0, h)

	termi.Clear()
	termi.HomeCursor()
	termi.Cooked()
	termi.ShowCursor()
}

func (f *FEP) drawStatus() {
	_, h := termi.Size()
	termi.SaveCursor()
	termi.HideCursor()
	termi.MoveCursor(0, h-1)

	status := f.status()
	termi.Print(status)
	termi.ClearTail()

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

		f.drawStatus()
	}
}
