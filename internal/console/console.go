package console

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

var state *term.State

func Raw() {
	if state != nil {
		term.Restore(int(os.Stdin.Fd()), state)
		state = nil
	}
	s, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	state = s
}

func Cooked() {
	if state == nil {
		panic("invalid state")
	}
	term.Restore(int(os.Stdin.Fd()), state)
}

func Clear() {
	fmt.Print("\x1b[2J")
}

func HomeCursor() {
	fmt.Print("\x1b[H")
}

func MoveCursor(x, y int) {
	fmt.Printf("\x1b[%d;%dH", y+1, x+1)
}

func HideCursor() {
	fmt.Print("\x1b[?25l")
}

func ShowCursor() {
	fmt.Print("\x1b[?25h")
}

func Size() (int, int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80, 24
	}
	return w, h
}

func EnableInvert() {
	fmt.Print("\x1b[7m")
}

func DisableInvert() {
	fmt.Print("\x1b[0m")
}

func SaveCursor() {
	fmt.Print("\x1b[s")
}

func LoadCursor() {
	fmt.Print("\x1b[u")
}

func ScrollRange(top, bottom int) {
	fmt.Printf("\x1b[%d;%dr", top+1, bottom)
}
