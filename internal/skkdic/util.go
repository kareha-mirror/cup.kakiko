package skkdic

import (
	"strings"
	"unicode/utf8"
)

// \\ \/ \n
func unescape(s string) string {
	buf := strings.Builder{}
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		if r != '\\' {
			buf.WriteRune(r)
			continue
		}
		r, size = utf8.DecodeRuneInString(s[i:])
		i += size
		if r == '\\' || r == '/' {
			buf.WriteRune(r)
			continue
		}
		if r == 'n' {
			buf.WriteRune('\n')
			continue
		}
		// undefined
		buf.WriteRune(r)
	}
	return buf.String()
}

func indexOfUnescapedSlash(s string) int {
	esc := false
	for i, r := range s {
		if esc {
			esc = false
		} else if r == '\\' {
			esc = true
		} else if r == '/' {
			return i
		}
	}
	return -1
}

func parseBody(line string) []string {
	if line == "" {
		return []string{}
	}
	line = strings.TrimSpace(line)

	defaultsRaw := make([]string, 0)
	blocksRaw := make([]string, 0)
	buf := strings.Builder{}
	inBr := false
	brBuf := strings.Builder{}
	esc := false

	flushDefault := func() {
		if buf.Len() > 0 {
			s := strings.TrimSpace(buf.String())
			if s != "" {
				defaultsRaw = append(defaultsRaw, s)
			}
			buf.Reset()
		}
	}

	for _, r := range line {
		if esc {
			if inBr {
				brBuf.WriteRune(r)
			} else {
				buf.WriteRune(r)
			}
			esc = false
		} else if r == '\\' {
			if inBr {
				brBuf.WriteRune(r)
			} else {
				buf.WriteRune(r)
			}
			esc = true
		} else if inBr {
			if r == ']' {
				blocksRaw = append(blocksRaw, brBuf.String())
				brBuf.Reset()
				inBr = false
			} else {
				brBuf.WriteRune(r)
			}
		} else {
			if r == '[' {
				inBr = true
				brBuf.Reset()
			} else if r == '/' {
				flushDefault()
			} else {
				buf.WriteRune(r)
			}
		}
	}
	flushDefault()

	defaults := make([]string, 0)
	for _, rawc := range defaultsRaw {
		segs := strings.Split(rawc, ";")
		if len(segs) < 1 {
			continue
		}
		surf := unescape(strings.TrimSpace(segs[0]))
		if surf == "" {
			continue
		}
		defaults = append(defaults, surf)
	}

	return defaults
}
