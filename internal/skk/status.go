package skk

import (
	"fmt"
	"strings"
)

func (en *Engine) Status() (string, bool) {
	// message
	if en.message != "" {
		return en.message, true
	}

	// list of candidates
	if en.conv.index >= candOffset {
		s := strings.Builder{}

		for i, r := range candKeyList {
			if en.conv.index+i >= len(en.conv.cands) {
				break
			}

			s.WriteRune(r + 'A' - 'a')
			s.WriteRune(':')

			cand := en.conv.candByIndex(en.conv.index + i)
			s.WriteString(cand)
			s.WriteString("  ")
		}

		k := max(len(en.conv.cands)-en.conv.index-len(candKeys), 0)
		s.WriteString(fmt.Sprintf("[残り %d]", k))

		return s.String(), false
	}

	// input mode
	s := strings.Builder{}
	s.WriteRune('(')
	switch en.inputMode {
	case inputASCII:
		s.WriteString("SKK")
	case inputHira:
		s.WriteString("かな")
	case inputKata:
		s.WriteString("カナ")
	case inputZen:
		s.WriteString("全英")
	default:
		s.WriteString("不明")
	}
	s.WriteRune(')')

	// registration buffer or line buffer
	if en.regMode {
		s.WriteString("[登録]")
		conv := en.convStack[len(en.convStack)-1]
		if conv.mode == convOkuri {
			stem := ""
			temp, ok := conv.stem.Substring(0, conv.stem.Len()-1)
			if ok {
				stem = temp
			}
			s.WriteString(stem)
			s.WriteRune('*')
		} else {
			s.WriteString(conv.stem.String())
		}
		s.WriteString(conv.okuri.String())
		s.WriteRune(' ')
		s.WriteString(en.conv.out.String())
		s.WriteString(en.regBuf.String())
	} else if en.lineMode {
		s.WriteString(en.lineBuf.String())
		s.WriteRune(':')
	}

	// stem
	if en.conv.mode != convNone {
		if en.conv.hasCands() {
			s.WriteRune('▼')
			s.WriteString(en.conv.cand())
		} else {
			s.WriteRune('▽')
			stem := ""
			if en.conv.mode == convOkuri {
				temp, ok := en.conv.stem.Substring(0, en.conv.stem.Len()-1)
				if ok {
					stem = temp
				}
			} else {
				stem = en.conv.stem.String()
			}
			s.WriteString(stem)
		}
	}

	// okuri
	if en.conv.mode == convOkuri && en.inputBuf.Len() > 0 {
		s.WriteRune('*')
	}
	s.WriteString(en.conv.okuri.String())

	// input
	s.WriteString(en.inputBuf.String())

	// pseudo cursor
	s.WriteRune('_')

	return s.String(), false
}
