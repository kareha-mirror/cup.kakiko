package fep

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tea.kareha.org/cup/termi"
	"tea.kareha.org/cup/termi/lock"
	"tea.kareha.org/cup/termi/pity"
)

const bufferSize = 1024

type Cmd int

const (
	CmdNone Cmd = iota
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
	dir   string
	cfg   *Config
	color termi.ColorPair

	p        *pity.Pity
	en       Engine
	listener termi.EscapeListener
	esc      bool
	outCh    chan []byte
	done     chan struct{}
}

func (fep *FEP) updateSize() error {
	w, h := termi.Size()
	return fep.p.SetSize(w, h-1)
}

func writeStringAll(p *pity.Pity, s string) error {
	data := []byte(s)
	total := 0

	for total < len(data) {
		n, err := p.Write(data[total:])
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

func Init(dir string, en Engine, cmd string, args ...string) (*FEP, error) {
	var cfg *Config
	cfgPath := getConfigPath(dir)
	_, err := os.Stat(cfgPath)
	if err != nil {
		cfg = DefaultConfig()
		SaveConfig(cfgPath, cfg)
	} else {
		cfg = LoadConfig(cfgPath)
	}

	color, err := termi.ParseColorPair(cfg.Color)
	if err != nil {
		return nil, err
	}

	termi.EscapeTimeout =
		time.Duration(cfg.EscapeTimeout) * time.Millisecond

	p, err := pity.Start(cmd, args...)
	if err != nil {
		return nil, err
	}

	fep := &FEP{
		dir:   dir,
		cfg:   cfg,
		color: color,

		p:        p,
		en:       en,
		listener: nil,
		esc:      false,
		outCh:    make(chan []byte, 128),
		done:     make(chan struct{}),
	}

	winch(fep)

	_, h := termi.Size()
	fmt.Print(termi.ScrollRange(0, h-1))

	fmt.Print(termi.Clear)
	fmt.Print(termi.HomeCursor)
	termi.Raw()
	termi.InitKey()

	err = lock.Lock(dir)
	if err != nil {
		reset()
		p.Close()
		return nil, err
	}
	err = en.Init(dir)
	lock.Unlock(dir)
	if err != nil {
		reset()
		p.Close()
		return nil, err
	}

	go func() {
		for data := range fep.outCh {
			os.Stdout.Write(data)
		}
	}()

	fep.draw()

	go func() {
		for {
			key := <-termi.Keys()
			if key.Kind == termi.KeyResized {
				err := fep.updateSize()
				if err != nil {
					return
				}
			}
			processed, cmd := en.Process(key)
			if processed != "" {
				err = writeStringAll(p, processed)
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
		p.Wait()
		close(fep.done)
		close(fep.outCh)
	}()

	listener := func(esc bool) {
		fep.esc = esc
		fep.draw()
	}
	fep.listener = termi.EscapeListener(&listener)
	termi.SetEscapeListener(fep.listener)

	return fep, nil
}

func reset() {
	termi.FinishKey()
	fmt.Print(termi.ScrollReset)
	fmt.Print(termi.Clear)
	fmt.Print(termi.HomeCursor)
	termi.Cooked()
	fmt.Print(termi.ShowCursor)
}

func (fep *FEP) Finish() error {
	err := lock.Lock(fep.dir)
	if err == nil {
		err = fep.en.Finish()
		lock.Unlock(fep.dir)
	}

	termi.SetEscapeListener(nil)
	reset()
	return err
}

func (fep *FEP) sync() error {
	err := lock.Lock(fep.dir)
	if err != nil {
		return err
	}
	err = fep.en.Sync()
	lock.Unlock(fep.dir)
	return err
}

func (fep *FEP) draw() {
	w, h := termi.Size()
	buf := strings.Builder{}

	buf.WriteString(termi.HideCursor)
	buf.WriteString(termi.SaveCursor)
	buf.WriteString(termi.MoveCursor(0, h-1))

	buf.WriteString(fep.color.Seq())

	status, inv := fep.en.Status()
	if inv {
		buf.WriteString(termi.SetInvert)
	}
	buf.WriteString(status)
	buf.WriteString(termi.ClearTail)
	if inv {
		buf.WriteString(termi.ResetInvert)
	}

	buf.WriteString(termi.MoveCursor(w-2, h-1))
	if fep.esc {
		buf.WriteString(" *")
	} else {
		buf.WriteString(" .")
	}

	buf.WriteString(termi.ResetAttr)

	buf.WriteString(termi.LoadCursor)
	buf.WriteString(termi.ShowCursor)

	data := []byte(buf.String())

	select {
	case fep.outCh <- data:
	case <-fep.done:
		return
	}
}

func (fep *FEP) Main() {
	buf := make([]byte, bufferSize)
	for {
		n, err := fep.p.Read(buf)
		if err != nil {
			return
		}

		data := append([]byte(nil), buf[:n]...)
		select {
		case fep.outCh <- data:
		case <-fep.done:
			return
		}
	}
}
