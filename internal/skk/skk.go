package skk

import (
	"fmt"

	"tea.kareha.org/lab/termi"
)

// XXX dummy
var count int = 0

func Process(key termi.Key) string {
	count++ // XXX dummy
	switch key.Kind {
	case termi.KeyRune:
		return string(key.Rune)
	default:
		return key.Raw
	}
}

func Status() string {
	// XXX dummy
	return fmt.Sprintf("Hello, World! (%d)", count)
}
