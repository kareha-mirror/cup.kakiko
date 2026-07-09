//go:build windows

package fep

func winch(fep *FEP) {
	fep.updateSize()
}
