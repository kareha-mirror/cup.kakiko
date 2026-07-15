package main

import "os"

func main() {
	for _, path := range os.Args[1:] {
		_ = os.RemoveAll(path)
	}
}
