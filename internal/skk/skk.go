package skk

import (
	"fmt"
)

// XXX dummy
var count int = 0

func Process(b []byte) []byte {
	count++ // XXX dummy
	return b
}

func Status() string {
	// XXX dummy
	return fmt.Sprintf("Hello, World! (%d)", count)
}
