//go:build !minimal

package main

import (
	_ "embed"
)

const appName = "kakiko"

//go:embed skk-edic-default.txt
var skkEdicDefault string
