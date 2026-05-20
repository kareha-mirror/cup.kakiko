package main

import (
	"fmt"
	"unicode"
)

func main() {
	for r := 0x3040; r <= 0x309f; r++ {
		k := r + 0x60
		fmt.Printf(
			"%v %04x %c %c %04x %v\n",
			unicode.In(rune(r), unicode.Hiragana),
			r, r, k, k,
			unicode.In(rune(k), unicode.Katakana),
		)
	}
}
