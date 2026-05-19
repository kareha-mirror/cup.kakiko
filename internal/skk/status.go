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

	// delete candidate
	if en.deleteMode {
		s := fmt.Sprintf(
			"%s /%s/ を辞書から削除します。良いですか？(y or n)",
			en.conv.stem.String(),
			en.conv.cand(),
		)
		return s, false
	}

	// list of candidates
	if en.conv.hasCands() && en.conv.index >= candOffset {
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

	s := strings.Builder{}

	// input mode
	s.WriteRune('(')
	switch en.inputMode {
	case inputASCII:
		s.WriteString("SKK:")
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

	// line buffer
	if en.lineMode && !en.regMode {
		s.WriteString(en.lineBuf.String())
		s.WriteRune(':')
	}

	// registration buffer
	if en.regMode {
		s.WriteString("[登録]")
		conv := en.stack[len(en.stack)-1]
		s.WriteString(conv.trueStem())
		if conv.mode == convOkuri {
			s.WriteRune('*')
		}
		s.WriteString(conv.okuri.String())
		s.WriteRune(' ')
		s.WriteString(en.conv.out.String())
		s.WriteString(en.regBuf.String())
	}

	// stem
	if en.conv.mode != convNone {
		if en.conv.hasCands() {
			s.WriteRune('▼')
			s.WriteString(en.conv.cand())
		} else {
			s.WriteRune('▽')
			s.WriteString(en.conv.trueStem())
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
