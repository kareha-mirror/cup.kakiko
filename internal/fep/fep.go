package fep

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
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
	outCh    chan []byte
	done     chan struct{}
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

func Unlock(dir string) error {
	path := getLockPath(dir)
	return os.Remove(path)
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

	termi.EscapeTimeout = time.Duration(cfg.EscTimeout) * time.Millisecond

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
		outCh:    make(chan []byte, 128),
		done:     make(chan struct{}),
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
	fmt.Print(termi.ScrollRange(0, h-1))

	fmt.Print(termi.Clear)
	fmt.Print(termi.HomeCursor)
	termi.Raw()
	termi.StartKey()

	err = lock(dir)
	if err != nil {
		reset()
		return nil, err
	}
	err = en.Init(dir)
	Unlock(dir)
	if err != nil {
		reset()
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
	termi.StopKey()
	fmt.Print(termi.ScrollReset)
	fmt.Print(termi.Clear)
	fmt.Print(termi.HomeCursor)
	termi.Cooked()
	fmt.Print(termi.ShowCursor)
}

func (fep *FEP) Finish() error {
	err := lock(fep.dir)
	if err == nil {
		err = fep.en.Finish()
		Unlock(fep.dir)
	}

	termi.SetEscapeListener(nil)
	reset()
	return err
}

func (fep *FEP) sync() error {
	err := lock(fep.dir)
	if err != nil {
		return err
	}
	err = fep.en.Sync()
	Unlock(fep.dir)
	return err
}

func (fep *FEP) draw() {
	w, h := termi.Size()
	buf := strings.Builder{}

	buf.WriteString(termi.HideCursor)
	buf.WriteString(termi.SaveCursor)
	buf.WriteString(termi.MoveCursor(0, h-1))

	buf.WriteString(fep.fgColor.Fg())
	buf.WriteString(fep.bgColor.Bg())

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
		n, err := fep.f.Read(buf)
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
