package console

import (
	"fmt"
)

func Print(s string) {
	fmt.Print(s)
}

func Printf(format string, a ...any) (n int, err error) {
	s := fmt.Sprintf(format, a...)
	Print(s)
	return len(s), nil
}
