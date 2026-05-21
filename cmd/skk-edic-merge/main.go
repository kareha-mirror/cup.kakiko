package main

import (
	"fmt"
	"os"

	"tea.kareha.org/cup/kakiko/internal/skkdic"
)

func fatal(a ...any) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func main() {
	m := map[string]string{}

	for _, path := range os.Args[1:] {
		r, err := os.Open(path)
		if err != nil {
			fatal(err)
		}

		err = skkdic.Load(r, m)
		if err != nil {
			fatal(err)
		}
	}

	skkdic.SaveSorted(os.Stdout, m)
}
