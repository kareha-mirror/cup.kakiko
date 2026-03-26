package main

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"

	"tea.kareha.org/lab/kakiko/internal/console"
)

var defaultCommand []string = []string{
	"/bin/sh",
}

func main() {
	var cmd []string
	if len(os.Args) > 1 {
		cmd = os.Args[1:]
	} else {
		cmd = defaultCommand
	}
	var c *exec.Cmd
	if len(cmd) < 2 {
		c = exec.Command(cmd[0])
	} else {
		c = exec.Command(cmd[0], cmd[1:]...)
	}

	ptmx, err := pty.Start(c)
	if err != nil {
		panic(err)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)

	go func() {
		for range ch {
			rows, cols, err := pty.Getsize(os.Stdin)
			if err != nil {
				panic(err)
			}
			pty.Setsize(ptmx, &pty.Winsize{
				Rows: uint16(rows - 1),
				Cols: uint16(cols),
			})
		}
	}()
	rows, cols, err := pty.Getsize(os.Stdin)
	if err != nil {
		panic(err)
	}
	pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(rows - 1),
		Cols: uint16(cols),
	})

	console.ScrollRange(0, rows-1)
	defer console.ScrollRange(0, rows)

	console.Clear()
	console.HomeCursor()
	console.Raw()
	defer func() {
		console.Clear()
		console.HomeCursor()
		console.Cooked()
		console.ShowCursor()
	}()

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

			processed := skkProcess(b)

			_, err = ptmx.Write(processed)
			if err != nil {
				return
			}
		}
	}()

	go func() {
		c.Wait()
		os.Exit(0)
	}()

	buf := make([]byte, 1024)
	for {
		n, err := ptmx.Read(buf)
		if err != nil {
			return
		}

		os.Stdout.Write(buf[:n])

		drawStatus()
	}
}

func drawStatus() {
	_, h := console.Size()
	console.SaveCursor()
	console.HideCursor()
	console.MoveCursor(0, h-1)
	console.Print("Hello, World!")
	console.ShowCursor()
	console.LoadCursor()
}

func skkProcess(b []byte) []byte {
	return b
}
