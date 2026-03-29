package main

import (
	"os"

	"tea.kareha.org/cup/kakiko/internal/fep"
	"tea.kareha.org/cup/kakiko/internal/skk"
)

func main() {
	f := fep.Init(os.Args, skk.Process, skk.Status)
	defer f.Finish()
	f.Main()
}
