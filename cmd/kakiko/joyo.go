//go:build joyo

package main

import (
	_ "embed"
)

const appName = "kakiko-joyo"

//go:embed skk-edic-joyo.txt
var skkdicBuiltin string
