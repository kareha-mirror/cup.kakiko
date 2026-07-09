//go:build unix

package fep

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

type Pty struct{
	c *exec.Cmd
	f *os.File
}

func PtyGetSize(in *os.File) (int, int, error) {
	rows, cols, err := pty.Getsize(in)
	return cols, rows, err
}

func (p *Pty) PtySetSize(w, h int) error {
	return pty.Setsize(p.f, &pty.Winsize{
		Rows: uint16(h),
		Cols: uint16(w),
	})
}

func PtyStart(cmd string, args ...string) (*Pty, error) {
	c := exec.Command(cmd, args...)
	f, err := pty.Start(c)
	if err != nil {
		return nil, err
	}
	return &Pty{
		c: c,
		f: f,
	}, nil
}

func (p *Pty) Wait() error {
	return p.c.Wait()
}
