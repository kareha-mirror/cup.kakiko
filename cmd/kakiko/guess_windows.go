//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func guessCommand() string {
	if p, err := exec.LookPath("pwsh.exe"); err == nil {
		return p
	}

	systemRoot := os.Getenv("SystemRoot")
	if systemRoot == "" {
		systemRoot = `C:\Windows`
	}

	ps := filepath.Join(
		systemRoot, "System32", "WindowsPowerShell", "v1.0", "powershell.exe",
	)
	if _, err := os.Stat(ps); err == nil {
		return ps
	}

	if comspec := os.Getenv("COMSPEC"); comspec != "" {
		return comspec
	}

	return filepath.Join(systemRoot, "System32", "cmd.exe")
}
