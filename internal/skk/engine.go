package skk

import (
	"fmt"
	"strings"

	"tea.kareha.org/cup/termi"

	"tea.kareha.org/cup/kakiko/internal/romaji"
)

type inputMode int

const (
	inputASCII inputMode = iota
	inputHira
	inputKata
	inputZen
)

type Engine struct {
	d Dics

	inputMode inputMode
	inputBuf  termi.RuneBuf

	conv      *convState
	convStack []*convState
	regMode   bool
	regBuf    termi.RuneBuf

	lineMode bool
	lineBuf  termi.RuneBuf
	linePass bool

	message string
}

func NewEngine(path string) *Engine {
	d := Dics{}
	d.SetUserDic(NewMemDic())
	d.AddDic(NewCDBDic(path))
	en := &Engine{
		d: d,

		inputMode: inputASCII,
		inputBuf:  termi.RuneBuf{},
		conv:      newConvState(),
		convStack: []*convState{},
		regMode:   false,
		regBuf:    termi.RuneBuf{},

		lineMode: false,
		lineBuf:  termi.RuneBuf{},
		linePass: false,

		message: "",
	}

	return en
}

func (en *Engine) pushConv() {
	en.convStack = append(en.convStack, en.conv)
	en.conv = newConvState()
}

func (en *Engine) popConv() bool {
	n := len(en.convStack)
	if n < 1 {
		return false
	}
	en.conv = en.convStack[n-1]
	en.convStack = en.convStack[:n-1]
	return n > 1
}

func (en *Engine) startReg() {
	en.regMode = true
	en.conv.out.WriteString(en.regBuf.String())
	en.regBuf.Reset()
	en.pushConv()
}

func (en *Engine) endReg() {
	en.inputBuf.Reset()
	if !en.popConv() {
		en.regMode = false
	}
}

func (en *Engine) flush(output *strings.Builder) {
	s := strings.Builder{}
	if en.conv.out.Len() > 0 {
		s.WriteString(en.conv.out.String())
	} else if en.conv.hasCands() {
		if en.conv.okuri.Len() > 0 {
			en.d.AddOkuri(
				en.conv.stem.String(),
				en.conv.okuri.String(),
				en.conv.cand(),
			)
		} else {
			en.d.Add(en.conv.stem.String(), en.conv.cand())
		}
		s.WriteString(en.conv.cand())
	} else {
		s.WriteString(en.conv.stem.String())
	}
	s.WriteString(en.conv.okuri.String())
	if en.regMode {
		en.regBuf.WriteString(s.String())
	} else if en.lineMode {
		en.lineBuf.WriteString(s.String())
	} else {
		output.WriteString(s.String())
	}
}

func (en *Engine) handleBackspace(r rune) (string, bool) {
	if en.inputBuf.Len() > 0 {
		en.inputBuf.Reset()
		return "", true
	}
	if en.conv.okuri.RemoveTail() {
		return "", true
	}
	if en.conv.mode != convNone && en.conv.stem.RemoveTail() {
		if en.conv.stem.Len() < 1 {
			en.conv.mode = convNone
		}
		en.conv.clearCands()
		return "", true
	}
	if en.regMode {
		if en.regBuf.RemoveTail() {
			return "", true
		}
		if en.conv.out.RemoveTail() {
			return "", true
		}
		en.endReg()
		return "", true
	}
	if en.lineMode && en.lineBuf.RemoveTail() {
		return "", true
	}
	return string(r), false
}

func (en *Engine) handleCancel(
	r rune, output *strings.Builder,
) (string, bool) {
	if en.regMode {
		if en.inputBuf.Len() > 0 {
			en.inputBuf.Reset()
			return output.String(), true
		}
		if en.regBuf.Len() > 0 || en.conv.out.Len() > 0 {
			en.regBuf.Reset()
			en.conv.out.Reset()
			return output.String(), true
		}
		en.endReg()
		return output.String(), true
	}

	switch en.conv.mode {
	default: //case convNone:
		if en.inputBuf.Len() > 0 {
			en.inputBuf.Reset()
			return output.String(), true
		}
		output.WriteRune(r)
		return output.String(), false
	case convOkuri:
		en.inputBuf.Reset()
		if !en.conv.hasCands() {
			en.conv.reset()
		} else {
			en.conv.stem.RemoveTail()
			en.conv.stem.WriteString(en.conv.okuri.String())
			en.conv.okuri.Reset()
			en.conv.clearCands()
			en.conv.mode = convStem
		}
		return output.String(), true
	case convStem, convAbbrev:
		en.inputBuf.Reset()
		if !en.conv.hasCands() {
			en.conv.reset()
		} else {
			en.conv.clearCands()
		}
		return output.String(), true
	}
}

func (en *Engine) handleEnter(r rune, output *strings.Builder) (string, bool) {
	reg := false
	if en.regMode {
		en.flush(output)
		en.endReg()

		regWord := en.regBuf.String()
		en.regBuf.Reset()

		if en.conv.mode == convOkuri {
			en.d.AddOkuri(
				en.conv.stem.String(), en.conv.okuri.String(), regWord,
			)
		} else {
			en.d.Add(en.conv.stem.String(), regWord)
		}

		en.conv.out.WriteString(regWord)
		en.conv.stem.Reset()
		en.conv.okuri.Reset()
		en.conv.mode = convNone
		if en.regMode {
			return output.String(), true
		}

		en.flush(output)
		en.conv.reset()
		reg = true
	}

	if en.lineMode {
		if en.lineBuf.Len() > 0 {
			output.WriteString(en.lineBuf.String())
			en.lineBuf.Reset()
		} else {
			output.WriteRune(r)
		}
		return output.String(), true
	} else {
		if !reg {
			output.WriteRune(r)
		}
		return output.String(), reg
	}
}

func (en *Engine) handleRune(r rune, output *strings.Builder) (string, bool) {
	if en.regMode {
		en.regBuf.WriteRune(r)
		return output.String(), true
	} else if en.lineMode {
		en.lineBuf.WriteRune(r)
		return output.String(), true
	} else {
		output.WriteRune(r)
		return output.String(), false
	}
}

func (en *Engine) handleConv(output *strings.Builder) (string, bool) {
	if !en.conv.hasCands() {
		if en.inputBuf.String() == "n" {
			if en.inputMode == inputHira {
				en.conv.stem.WriteString("ん")
			} else { // inputKata
				en.conv.stem.WriteString("ン")
			}
		}
		en.inputBuf.Reset()

		stem := en.conv.stem.String()
		stem = romaji.KataToHira(stem)
		var err error
		if en.conv.mode == convOkuri {
			okuri := en.conv.okuri.String()
			okuri = romaji.KataToHira(okuri)
			en.conv.cands, err = en.d.LookupOkuri(stem, okuri)
		} else {
			en.conv.cands, err = en.d.Lookup(stem)
		}
		en.conv.index = 0
		if err != nil {
			en.message = fmt.Sprintf("%v", err)
			en.conv.cands = []string{}
		} else if !en.conv.hasCands() {
			en.startReg()
			return output.String(), true
		}
	} else {
		if en.conv.index < candOffset {
			if en.conv.index+1 < len(en.conv.cands) {
				en.conv.index++
			} else {
				en.startReg()
				return output.String(), true
			}
		} else {
			if en.conv.index+len(candKeys) < len(en.conv.cands) {
				en.conv.index += len(candKeys)
			} else {
				en.startReg()
				return output.String(), true
			}
		}
	}
	return output.String(), true
}

func (en *Engine) handleConvRev(output *strings.Builder) (string, bool) {
	if en.conv.index > candOffset {
		en.conv.index -= len(candKeys)
	} else {
		en.conv.index--
	}
	if en.conv.index < 0 {
		en.conv.okuri.Reset()
		en.conv.clearCands()
	}
	return output.String(), true
}

func vowelOf(kana string) (string, bool) {
	switch kana {
	case "あ":
		return "a", true
	case "い":
		return "i", true
	case "う":
		return "u", true
	case "え":
		return "e", true
	case "お":
		return "o", true
	default:
		return "", false
	}
}
