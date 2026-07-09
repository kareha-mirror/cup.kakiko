//go:build windows

package fep

import (
	"context"
	"os"

	"github.com/UserExistsError/conpty"
	"tea.kareha.org/cup/termi"
)

type Pty struct{
	f *conpty.ConPty
}

func PtyGetSize(in *os.File) (int, int, error) {
	w, h := termi.Size()
	return w, h, nil
}

func (p *Pty) PtySetSize(w, h int) error {
	return p.f.Resize(w, h)
}

func PtyStart(cmd string, args ...string) (*Pty, error) {
	f, err := conpty.Start(cmd) // XXX args?
	if err != nil {
		return nil, err
	}
	return &Pty{
		f: f,
	}, nil
}

func (p *Pty) Wait() error {
	/*exitCode*/ _, err := p.f.Wait(context.Background())
	return err
}
