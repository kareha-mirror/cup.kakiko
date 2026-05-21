package skkdic

import (
	"strings"
)

func parseSeq(seq string) []string {
	cands := make([]string, 0)
	buf := strings.Builder{}
	esc := false

	flush := func() {
		if buf.Len() > 0 {
			cands = append(cands, buf.String())
			buf.Reset()
		}
	}

	for _, r := range seq {
		if esc {
			if r == '\\' || r == '/' {
				buf.WriteRune(r)
			}
			esc = false
		} else if r == '\\' {
			esc = true
		} else {
			if r == '/' {
				flush()
			} else {
				buf.WriteRune(r)
			}
		}
	}
	flush()

	return cands
}

func escape(s string) string {
	buf := []rune{}
	for _, r := range s {
		if r == '\\' || r == '/' {
			buf = append(buf, '\\')
		}
		buf = append(buf, r)
	}
	return string(buf)
}
