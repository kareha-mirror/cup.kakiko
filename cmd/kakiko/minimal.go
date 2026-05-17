//go:build minimal

package main

import (
	_ "embed"
)

const appName = "kakikom"

//go:embed skk-edic-joyo.txt
var skkEdicDefault string
