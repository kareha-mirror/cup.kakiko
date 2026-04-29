package skk

import (
	"fmt"
	"strings"

	"tea.kareha.org/cup/kakiko/internal/romaji"
)

func (en *Engine) enterZenMode() (string, bool) {
	en.inputMode = inputZen
	en.inputBuf.Reset()

	if en.conv.mode != convNone {
		en.flush()
		en.conv.reset()
	}

	return en.output(true)
}

func (en *Engine) handleConvEnter() (string, bool) {
	if en.regMode {
		en.flush()
		en.inputBuf.Reset()
		if !en.popConv() {
			en.regMode = false
		}

		regWord := en.regBuf.String()
		en.regBuf.Reset()

		if en.conv.mode == convOkuri {
			en.dics.AddOkuri(
				en.conv.stem.String(), en.conv.okuri.String(), regWord,
			)
		} else {
			en.dics.Add(en.conv.stem.String(), regWord)
		}

		en.conv.out.WriteString(regWord)
		en.conv.stem.Reset()
		en.conv.okuri.Reset()
		en.conv.mode = convNone
		if en.regMode {
			return en.output(true)
		}
	}

	en.flush()
	en.conv.reset()
	return en.output(true)
}

func (en *Engine) handleEscape(r rune) (string, bool) {
	en.inputMode = inputASCII
	en.inputBuf.Reset()

	if en.conv.mode != convNone {
		en.flush()
		en.conv.reset()
	}

	// XXX regMode?

	if en.lineMode {
		en.out.WriteString(en.lineBuf.String())
		en.lineBuf.Reset()
		en.lineMode = false
	}

	en.out.WriteRune(r)
	return en.output(true)
}

func (en *Engine) handleControlCode(r rune) (string, bool) {
	en.out.WriteRune(r)
	return en.output(false)
}

func (en *Engine) enterAbbrevMode() (string, bool) {
	if en.conv.hasCands() {
		en.flush()
		en.conv.reset()
	}
	en.conv.mode = convAbbrev
	return en.output(true)
}

func (en *Engine) handleConv() (string, bool) {
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
			en.conv.cands, err = en.dics.LookupOkuri(stem, okuri)
		} else {
			en.conv.cands, err = en.dics.Lookup(stem)
		}
		en.conv.index = 0
		if err != nil {
			en.message = fmt.Sprintf("%v", err)
			en.conv.cands = []string{}
		} else if !en.conv.hasCands() {
			en.beginReg()
		}
	} else {
		if en.conv.index < candOffset {
			if en.conv.index+1 < len(en.conv.cands) {
				en.conv.index++
			} else {
				en.beginReg()
			}
		} else {
			if en.conv.index+len(candKeys) < len(en.conv.cands) {
				en.conv.index += len(candKeys)
			} else {
				en.beginReg()
			}
		}
	}
	return en.output(true)
}

func (en *Engine) handleConvRev() (string, bool) {
	if en.conv.index > candOffset {
		en.conv.index -= len(candKeys)
	} else {
		en.conv.index--
	}
	if en.conv.index < 0 {
		en.conv.okuri.Reset()
		en.conv.clearCands()
	}
	return en.output(true)
}

func (en *Engine) handleKigou(kigou string, update bool) (string, bool) {
	if en.conv.mode != convNone && kigou == "ー" {
		if !en.conv.hasCands() {
			en.conv.stem.WriteString(kigou)
			return en.output(true)
		}
	}

	update = update || en.inputBuf.Len() > 0
	en.inputBuf.Reset()
	if en.regMode {
		en.regBuf.WriteString(kigou)
		update = true
	} else if en.lineMode {
		en.lineBuf.WriteString(kigou)
		update = true
	} else {
		en.out.WriteString(kigou)
	}
	return en.output(update)
}

func (en *Engine) handleNonAlpha(r rune, update bool) (string, bool) {
	if en.conv.mode != convNone {
		en.flush()
		en.conv.reset()
		update = true
	}

	if en.regMode {
		en.regBuf.WriteRune(r)
		update = true
	} else if en.lineMode {
		en.lineBuf.WriteRune(r)
		update = true
	} else {
		en.out.WriteRune(r)
	}
	return en.output(update)
}

func (en *Engine) enterASCIIMode() (string, bool) {
	en.inputMode = inputASCII
	en.inputBuf.Reset()

	if en.conv.mode != convNone {
		en.flush()
		en.conv.reset()
	}

	return en.output(true)
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

func (en *Engine) changeKanaType() (string, bool) {
	if en.conv.mode == convNone {
		if en.inputMode == inputHira {
			en.inputMode = inputKata
		} else { // inputKata
			en.inputMode = inputHira
		}
		return en.output(true)
	} else {
		if en.inputBuf.String() == "n" {
			if en.inputMode == inputHira {
				en.conv.stem.WriteString("ん")
			} else { // inputKata
				en.conv.stem.WriteString("ン")
			}
		}
		en.inputBuf.Reset()

		en.conv.mode = convNone
		s := strings.Builder{}
		if en.inputMode == inputHira {
			s.WriteString(romaji.HiraToKata(en.conv.stem.String()))
			s.WriteString(romaji.HiraToKata(en.conv.okuri.String()))
		} else { // inputKata
			s.WriteString(romaji.KataToHira(en.conv.stem.String()))
			s.WriteString(romaji.KataToHira(en.conv.okuri.String()))
		}
		en.conv.reset()
		if en.regMode {
			en.regBuf.WriteString(s.String())
		} else if en.lineMode {
			en.lineBuf.WriteString(s.String())
		} else {
			en.out.WriteString(s.String())
		}
		return en.output(true)
	}
}

func (en *Engine) handleAlpha(r rune, update bool) (string, bool) {
	en.inputBuf.WriteRune(r)

	var kana string
	hold := false
	if _, ok := romaji.IsSokuon[en.inputBuf.String()]; ok {
		if en.inputMode == inputHira {
			kana = "っ"
		} else { // inputKata
			kana = "ッ"
		}
		en.inputBuf.RemoveHead()
		hold = true
	} else if _, ok := romaji.IsN[en.inputBuf.String()]; ok {
		if en.inputMode == inputHira {
			kana = "ん"
		} else { // inputKata
			kana = "ン"
		}
		en.inputBuf.RemoveHead()
		hold = true
	} else {
		lookup := en.inputBuf.String()
		alias, ok := romaji.Aliases[lookup]
		if ok {
			lookup = alias
		}

		var k string
		if en.inputMode == inputHira {
			k, ok = romaji.ToHira[lookup]
		} else { // inputKata
			k, ok = romaji.ToKata[lookup]
		}
		if ok {
			kana = k
			en.inputBuf.Reset()
		}
	}

	if en.conv.mode == convNone {
		if kana != "" {
			if en.regMode {
				en.regBuf.WriteString(kana)
			} else if en.lineMode {
				en.lineBuf.WriteString(kana)
			} else {
				en.out.WriteString(kana)
			}
		}
		return en.output(true)
	} else if en.conv.mode == convStem {
		if kana != "" {
			en.conv.stem.WriteString(kana)
			en.conv.clearCands()
		}
		return en.output(true)
	} else if en.conv.mode == convOkuri {
		vowel, ok := vowelOf(kana)
		if ok {
			en.conv.stem.WriteString(vowel)
			en.conv.okuri.WriteString(kana)
		} else if kana != "" {
			en.conv.okuri.WriteString(kana)
		} else {
			en.conv.stem.WriteRune(r)
		}

		if en.conv.okuri.Len() < 1 {
			return en.output(true)
		}

		if hold {
			return en.output(true)
		}

		stem := en.conv.stem.String()
		stem = romaji.KataToHira(stem)
		okuri := en.conv.okuri.String()
		okuri = romaji.KataToHira(okuri)
		var err error
		en.conv.cands, err = en.dics.LookupOkuri(stem, okuri)
		en.conv.index = 0
		if err != nil {
			en.message = fmt.Sprintf("%v", err)
			en.conv.cands = []string{}
		} else if !en.conv.hasCands() {
			en.beginReg()
		}
		return en.output(true)
	} else {
		en.message = fmt.Sprintf(
			"Process: invalid conv.mode == %d", en.conv.mode,
		)
		return en.output(false)
	}
}
