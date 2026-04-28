package skk

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

type Dic interface {
	Lookup(reading string) ([]string, error)
	LookupOkuri(key, okuri string) ([]string, error)
}

type UserDic interface {
	Dic
	Add(reading, kanji string)
	AddOkuri(key, okuri, kanji string)
}

// \\ \/ \; \[ \] \n \t \" \' \u{...}
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
		if r == '\\' || r == '/' || r == ';' || r == '[' || r == ']' ||
			r == '\'' || r == '"' {
			buf.WriteRune(r)
			continue
		}
		if r == 'n' {
			buf.WriteRune('\n')
			continue
		}
		if r == 't' {
			buf.WriteRune('\t')
			continue
		}
		if r == 'u' {
			r, size = utf8.DecodeRuneInString(s[i:])
			if r != '{' {
				buf.WriteRune('u')
				continue
			}
			i += size
			k := strings.Index(s[i:], "}")
			if k < 0 {
				buf.WriteString("\\u{")
				continue
			}
			hex := s[i : i+k]
			n, err := strconv.ParseUint(hex, 16, 32)
			i += k + 1
			if err != nil {
				buf.WriteString("\\u{")
				buf.WriteString(hex)
				buf.WriteRune('}')
				continue
			}
			buf.WriteRune(rune(n))
			continue
		}
		// undefined
		buf.WriteRune(r)
	}
	return buf.String()
}

func splitSemicolon(s string) []string {
	fields := make([]string, 0)
	buf := strings.Builder{}
	esc := false
	for _, r := range s {
		if esc {
			buf.WriteRune(r)
			esc = false
		} else if r == '\\' {
			buf.WriteRune(r)
			esc = true
		} else if r == ';' {
			fields = append(fields, buf.String())
			buf.Reset()
		} else {
			buf.WriteRune(r)
		}
	}
	return fields
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

func parseBody(line string) ([]string, map[string][]string) {
	if line == "" {
		return []string{}, map[string][]string{}
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

	blocks := make(map[string][]string, 0)
	for _, br := range blocksRaw {
		br = strings.TrimSpace(br)
		pos := indexOfUnescapedSlash(br)
		if pos >= 0 {
			okuri := strings.TrimSpace(br[:pos])
			rest := br[pos+1:]

			toks := make([]string, 0)
			bbuf := strings.Builder{}
			esc2 := false
			for _, r := range rest {
				if esc2 {
					bbuf.WriteRune(r)
					esc2 = false
				} else if r == '\\' {
					bbuf.WriteRune(r)
					esc2 = true
				} else if r == '/' {
					s := strings.TrimSpace(bbuf.String())
					bbuf.Reset()
					if s == "" {
						continue
					}
					toks = append(toks, s)
				} else {
					bbuf.WriteRune(r)
				}
			}
			last := strings.TrimSpace(bbuf.String())
			if last != "" {
				toks = append(toks, last)
			}

			if okuri != "" && len(toks) > 0 {
				arr := make([]string, 0)
				for _, rawc := range toks {
					segs := splitSemicolon(rawc)
					if len(segs) < 1 {
						continue
					}
					surf := unescape(strings.TrimSpace(segs[0]))
					if surf == "" {
						continue
					}
					arr = append(arr, surf)
				}
				if len(arr) > 0 {
					blocks[okuri] = arr
				}
			}
		}
	}

	return defaults, blocks
}
