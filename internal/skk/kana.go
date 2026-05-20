package skk

import (
	"fmt"
	"strings"

	"tea.kareha.org/cup/kakiko/internal/romaji"
)

func (en *Engine) resetKanaInput() {
	if en.inputBuf.String() == "n" {
		nn := ""
		if en.inputMode == inputHira {
			nn = "ん"
		} else if en.inputMode == inputKata {
			nn = "ン"
		}
		if en.conv.mode == convOkuri {
			en.conv.stem.RemoveTail()
			en.conv.stem.WriteString(nn)
			en.conv.stem.WriteRune('n')
		} else if en.conv.mode == convStem || en.regMode {
			en.conv.stem.WriteString(nn)
		} else if en.conv.mode == convNone {
			en.conv.out.WriteString(nn)
		}
	}
	en.inputBuf.Reset()
}

func (en *Engine) enterZenMode() (string, bool) {
	en.resetKanaInput()
	en.inputMode = inputZen

	en.flush()
	en.resetConv()
	return en.output(true)
}

func (en *Engine) handleConvEnter() (string, bool) {
	if en.regMode {
		en.flush()
		en.endReg()

		word := en.regBuf.String()
		en.regBuf.Reset()

		en.conv.out.WriteString(word)
		if word != "" {
			en.dics.Add(en.conv.stem.String(), word)
			en.dics.AddDiff(en.conv.stem.String(), word)
		}

		en.conv.out.WriteString(en.conv.okuri.String())
		en.conv.stem.Reset()
		en.conv.okuri.Reset()
		en.conv.mode = convNone

		en.flushPartial()
		en.conv.reset() // do not use en.resetConv()
		return en.output(true)
	}

	en.flush()
	en.resetConv()
	return en.output(true)
}

func (en *Engine) handleEscape(r rune) (string, bool) {
	en.resetKanaInput()
	en.inputMode = inputASCII

	en.flush()
	en.resetConv()

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
	en.flush()
	en.resetConv()

	en.conv.mode = convAbbrev
	return en.output(true)
}

func (en *Engine) handleConv() (string, bool) {
	if en.conv.hasCands() {
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
	} else {
		en.resetKanaInput()

		var err error
		stem := romaji.KataToHira(en.conv.stem.String())
		en.conv.cands, err = en.dics.Lookup(stem)
		en.conv.index = 0
		if err != nil {
			en.message = fmt.Sprintf("%v", err)
			en.conv.cands = []string{}
		} else if !en.conv.hasCands() {
			en.beginReg()
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
	en.writeString(kigou)
	if en.regMode || en.lineMode {
		update = true
	}
	return en.output(update)
}

func (en *Engine) handleNonAlphabet(r rune) (string, bool) {
	en.resetKanaInput()

	switch en.conv.mode {
	case convOkuri:
		en.conv.okuri.WriteRune(r)
	case convStem:
		en.conv.stem.WriteRune(r)
	default:
		en.conv.out.WriteRune(r)
	}

	if en.conv.mode == convNone {
		en.flush()
		en.resetConv()
	}

	return en.output(true)
}

func (en *Engine) enterASCIIMode() (string, bool) {
	en.resetKanaInput()
	en.inputMode = inputASCII

	en.flush()
	en.resetConv()
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

func (en *Engine) toggleKanaType() (string, bool) {
	if en.conv.mode == convNone {
		if en.inputMode == inputHira {
			en.inputMode = inputKata
		} else { // inputKata
			en.inputMode = inputHira
		}
		return en.output(true)
	}

	en.resetKanaInput()

	en.conv.mode = convNone
	s := strings.Builder{}
	if en.inputMode == inputHira {
		s.WriteString(romaji.HiraToKata(en.conv.stem.String()))
		s.WriteString(romaji.HiraToKata(en.conv.okuri.String()))
	} else { // inputKata
		s.WriteString(romaji.KataToHira(en.conv.stem.String()))
		s.WriteString(romaji.KataToHira(en.conv.okuri.String()))
	}
	en.resetConv()
	en.writeString(s.String())
	return en.output(true)
}

func (en *Engine) handleAlphabet(r rune, update bool) (string, bool) {
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

	switch en.conv.mode {
	case convNone:
		if kana != "" {
			en.writeString(kana)
		}
		return en.output(true)
	case convStem:
		if kana != "" {
			en.conv.stem.WriteString(kana)
			en.conv.clearCands()
		}
		return en.output(true)
	case convOkuri:
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
		var err error
		en.conv.cands, err = en.dics.Lookup(stem)
		en.conv.index = 0
		if err != nil {
			en.message = fmt.Sprintf("%v", err)
			en.conv.cands = []string{}
		} else if !en.conv.hasCands() {
			en.beginReg()
		}
		return en.output(true)
	default:
		en.message = fmt.Sprintf(
			"Process: invalid conv.mode == %d", en.conv.mode,
		)
		return en.output(false)
	}
}
