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

	if r == '\a' { // Ctrl-G
		return en.handleCancel(r)
	}

	if en.conv.index >= candOffset && r != ' ' && r != 'x' {
		return en.handleCandList(r)
	}

	if r == '\f' { // Ctrl-L
		return en.handleLineMode(r)
	}
	en.linePass = false

	if r == '\n' { // Ctrl-J
		return en.handleCommit(r)
	}

	if r == termi.RuneEnter && en.conv.mode == convNone {
		return en.handleEnter(r)
	}

	if en.inputMode == inputZen {
		return en.handleZen(r)
	}

	if en.inputMode == inputASCII {
		return en.handleRune(r)
	}

	// now in Hiragana or Katakana mode
	// assert
	if en.inputMode != inputHira && en.inputMode != inputKata {
		en.message = fmt.Sprintf(
			"assert: invalid inputMode == %d", en.inputMode,
		)
		return en.output(true)
	}

	if r == 'L' {
		return en.handleZenMode()
	}

	if r == termi.RuneEnter && en.conv.mode != convNone {
		return en.handleKanaEnter()
	}

	if r == termi.RuneEscape {
		return en.handleEscape(r)
	}

	if r < ' ' {
		return en.handleControlCode(r)
	}

	if r == ' ' && en.conv.mode != convNone {
		return en.handleConv()
	}

	if r == 'x' && en.conv.hasCands() {
		return en.handleConvRev()
	}

	if r == '/' {
		return en.handleAbbrevMode()
	}

	if en.conv.mode == convAbbrev {
		if en.conv.hasCands() {
			en.flush()
			en.conv.reset()
		} else {
			en.conv.stem.WriteRune(r)
			return en.output(true)
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
					en.out.WriteString(en.conv.cand())
					en.out.WriteString(okuri)
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
				en.out.WriteString(en.conv.cand())
				en.out.WriteString(en.conv.okuri.String())
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
		en.conv.advanceMode()
		r += 'a' - 'A'
	}

	kigou, ok := romaji.ToKigou[string(r)]
	if ok {
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

	if r < 'a' || r > 'z' {
		return en.handleNonAlpha(r, update)
	}

	if r == 'l' {
		return en.handleASCIIMode()
	}

	if r == 'q' {
		if en.conv.mode == convNone {
			if en.inputMode == inputHira {
				en.inputMode = inputKata
			} else { // inputKata
				en.inputMode = inputHira
			}
			return en.output(true)
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
				en.out.WriteString(s)
			}
			return en.output(true)
		}
	}

	en.inputBuf.WriteRune(r)

	return en.handleKana(r, update)
}
