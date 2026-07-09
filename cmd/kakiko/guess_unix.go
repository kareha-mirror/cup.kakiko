//go:build unix

package main

import (
	"os"
)

const fallbackCommand = "/bin/sh"

func guessCommand() string {
	command := os.Getenv("SHELL")
	if command == "" {
		command = fallbackCommand
	}
	return command
}
