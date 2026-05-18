//go:build !joyo

package main

import (
	_ "embed"
)

const appName = "kakiko"

//go:embed skk-edic-default.txt
var skkdicBuiltin string
