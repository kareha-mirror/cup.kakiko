//go:build unix

package fep

import (
	"os"
	"os/signal"
	"syscall"
)

func winch(fep *FEP) {
	err := fep.updateSize()
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
}
