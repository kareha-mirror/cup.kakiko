package main

import (
	"fmt"
)

func main() {
	for r := 0x3040; r <= 0x309f; r++ {
		k := r + 0x60
		fmt.Printf("%04x %c %c %04x\n", r, r, k, k)
	}
}
