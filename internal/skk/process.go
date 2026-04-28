package skk

import (
	"fmt"
	"strings"

	"tea.kareha.org/cup/termi"

	"tea.kareha.org/cup/kakiko/internal/romaji"
)

func (en *Engine) Process(key termi.Key) (string, bool) {
	if en.message != "" {
		en.message = ""
		return "", true
	}

	if key.Kind != termi.KeyRune {
		return key.Raw, false
	}

	r := key.Rune

	if r == termi.RuneBackspace || r == termi.RuneDelete {
		return en.handleBackspace(r)
	}

	output := strings.Builder{}

	if r == '\a' { // Ctrl-G
		return en.handleCancel(r, &output)
	}

	if en.conv.index >= candOffset && r != ' ' && r != 'x' {
		index := en.conv.keyToIndex(r)
		if index < 0 {
			en.message = fmt.Sprintf(
				//"\"%c\" is not valid here!", r,
				"\"%c\" は有効なキーではありません！", r,
			)
			return output.String(), true
		}
		en.conv.index = index

		en.flush(&output)
		en.conv.reset()
		return output.String(), true
	}

	if r == '\f' { // Ctrl-L
		if en.regMode {
			output.WriteRune(r)
			return output.String(), true
		}

		if en.linePass {
			en.linePass = false
			output.WriteRune(r)
		} else {
			en.linePass = true
		}

		if en.lineMode {
			en.lineMode = false
			en.flush(&output)
			output.WriteString(en.lineBuf.String())
			en.lineBuf.Reset()
		} else {
			en.lineMode = true
		}

		return output.String(), true
	}
	en.linePass = false

	if r == '\n' { // Ctrl-J
		if en.conv.mode == convAbbrev {
			output.WriteString(romaji.HanToZen(en.conv.stem.String()))
			en.conv.reset()
			return output.String(), true
		}

		if en.conv.mode != convNone {
			en.flush(&output)
		}

		if en.inputMode != inputHira && en.inputMode != inputKata {
			en.inputMode = inputHira
		}

		en.conv.reset()

		return output.String(), true
	}

	if r == termi.RuneEnter && en.conv.mode == convNone {
		return en.handleEnter(r, &output)
	}

	if en.inputMode == inputZen {
		zen, ok := romaji.ToZen[string(r)]
		if ok {
			if en.regMode {
				en.regBuf.WriteString(zen)
				return output.String(), true
			} else if en.lineMode {
				en.lineBuf.WriteString(zen)
				return output.String(), true
			} else {
				output.WriteString(zen)
				return output.String(), false
			}
		} else {
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
	}

	if en.inputMode == inputASCII {
		return en.handleRune(r, &output)
	}

	// now in Hiragana or Katakana mode
	// assert
	if en.inputMode != inputHira && en.inputMode != inputKata {
		en.message = fmt.Sprintf(
			"assert: invalid inputMode == %d", en.inputMode,
		)
		return output.String(), true
	}

	if r == 'L' {
		en.inputMode = inputZen
		en.inputBuf.Reset()

		if en.conv.mode != convNone {
			en.flush(&output)
			en.conv.reset()
		}

		return output.String(), true
	}

	if r == termi.RuneEnter && en.conv.mode != convNone {
		if en.regMode {
			en.flush(&output)
			en.inputBuf.Reset()
			if !en.popConv() {
				en.regMode = false
			}

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
		}

		en.flush(&output)
		en.conv.reset()
		return output.String(), true
	}

	if r == termi.RuneEscape {
		en.inputMode = inputASCII
		en.inputBuf.Reset()

		if en.conv.mode != convNone {
			en.flush(&output)
			en.conv.reset()
		}

		// XXX regMode?

		if en.lineMode {
			output.WriteString(en.lineBuf.String())
			en.lineBuf.Reset()
			en.lineMode = false
		}

		output.WriteRune(r)
		return output.String(), true
	}

	if r < ' ' {
		output.WriteRune(r)
		return output.String(), false
	}

	if r == ' ' && en.conv.mode != convNone {
		return en.handleConv(&output)
	}

	if r == 'x' && en.conv.hasCands() {
		return en.handleConvRev(&output)
	}

	if r == '/' {
		if en.conv.hasCands() {
			en.flush(&output)
			en.conv.reset()
		}
		en.conv.mode = convAbbrev
		return output.String(), true
	}

	if en.conv.mode == convAbbrev {
		if en.conv.hasCands() {
			en.flush(&output)
			en.conv.reset()
		} else {
			en.conv.stem.WriteRune(r)
			return output.String(), true
		}
	}

	update := false
	if en.conv.hasCands() {
		if en.conv.mode == convOkuri {
			okuri := en.conv.okuri.String()
			if en.conv.okuri.Len() > 0 &&
				!strings.HasSuffix(okuri, "っ") &&
				!strings.HasSuffix(okuri, "ッ") {
				if en.regMode {
					en.regBuf.WriteString(en.conv.cand())
					en.regBuf.WriteString(okuri)
				} else if en.lineMode {
					en.lineBuf.WriteString(en.conv.cand())
					en.lineBuf.WriteString(okuri)
				} else {
					output.WriteString(en.conv.cand())
					output.WriteString(okuri)
				}
				en.conv.reset()
				en.inputBuf.Reset()
				update = true
			}
		} else {
			if en.regMode {
				en.regBuf.WriteString(en.conv.cand())
				en.regBuf.WriteString(en.conv.okuri.String())
			} else if en.lineMode {
				en.lineBuf.WriteString(en.conv.cand())
				en.lineBuf.WriteString(en.conv.okuri.String())
			} else {
				output.WriteString(en.conv.cand())
				output.WriteString(en.conv.okuri.String())
			}
			en.conv.reset()
			update = true
		}
	}

	if r >= 'A' && r <= 'Z' {
		if en.inputBuf.String() == "n" {
			if en.inputMode == inputHira {
				en.conv.stem.WriteString("ん")
			} else { // inputKata
				en.conv.stem.WriteString("ン")
			}
			en.inputBuf.Reset()
		}
		en.conv.advanceModeOnUpper()
		r += 'a' - 'A'
	}

	kigou, ok := romaji.ToKigou[string(r)]
	if ok {
		if en.conv.mode != convNone && kigou == "ー" {
			if !en.conv.hasCands() {
				en.conv.stem.WriteString(kigou)
				return output.String(), true
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
			output.WriteString(kigou)
		}
		return output.String(), update
	}

	if r < 'a' || r > 'z' {
		if en.conv.mode != convNone {
			en.flush(&output)
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
			output.WriteRune(r)
		}
		return output.String(), update
	}

	if r == 'l' {
		en.inputMode = inputASCII
		en.inputBuf.Reset()

		if en.conv.mode != convNone {
			en.flush(&output)
			en.conv.reset()
		}

		return output.String(), true
	}

	if r == 'q' {
		if en.conv.mode == convNone {
			if en.inputMode == inputHira {
				en.inputMode = inputKata
			} else { // inputKata
				en.inputMode = inputHira
			}
			return output.String(), true
		} else if en.conv.mode == convStem {
			if en.inputBuf.String() == "n" {
				if en.inputMode == inputHira {
					en.conv.stem.WriteString("ん")
				} else { // inputKata
					en.conv.stem.WriteString("ン")
				}
			}
			en.inputBuf.Reset()

			en.conv.mode = convNone
			var s string
			if en.inputMode == inputHira {
				s = romaji.HiraToKata(en.conv.stem.String())
			} else { // inputKata
				s = romaji.KataToHira(en.conv.stem.String())
			}
			en.conv.stem.Reset()
			if en.regMode {
				en.regBuf.WriteString(s)
			} else if en.lineMode {
				en.lineBuf.WriteString(s)
			} else {
				output.WriteString(s)
			}
			return output.String(), true
		}
	}

	en.inputBuf.WriteRune(r)

	var kana string
	sokuon := false
	if _, ok := romaji.IsSokuon[en.inputBuf.String()]; ok {
		if en.inputMode == inputHira {
			kana = "っ"
		} else { // inputKata
			kana = "ッ"
		}
		en.inputBuf.RemoveHead()
		sokuon = true
	} else if _, ok := romaji.IsN[en.inputBuf.String()]; ok {
		if en.inputMode == inputHira {
			kana = "ん"
		} else { // inputKata
			kana = "ン"
		}
		en.inputBuf.RemoveHead()
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
				output.WriteString(kana)
			}
		}
		return output.String(), true
	} else if en.conv.mode == convStem {
		if kana != "" {
			en.conv.stem.WriteString(kana)
			en.conv.clearCands()
		}
		return output.String(), true
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
			return output.String(), true
		}

		if sokuon {
			return output.String(), true
		}

		stem := en.conv.stem.String()
		stem = romaji.KataToHira(stem)
		okuri := en.conv.okuri.String()
		okuri = romaji.KataToHira(okuri)
		var err error
		en.conv.cands, err = en.d.LookupOkuri(stem, okuri)
		en.conv.index = 0
		if err != nil {
			en.message = fmt.Sprintf("%v", err)
		}
		return output.String(), true
	} else {
		en.message = fmt.Sprintf(
			"Process: invalid conv.mode == %d", en.conv.mode,
		)
		return output.String(), false
	}
}
