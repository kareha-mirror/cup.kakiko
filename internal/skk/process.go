package skk

import (
	"fmt"

	"tea.kareha.org/cup/termi"

	"tea.kareha.org/cup/kakiko/internal/romaji"
)

func (en *Engine) Process(key termi.Key) (string, bool) {
	// hide message
	if en.message != "" {
		en.message = ""
		return "", true
	}

	// pass-through
	if key.Kind != termi.KeyRune {
		return key.Raw, false
	}

	// shortcut
	r := key.Rune

	// backspace
	if r == termi.RuneBackspace || r == termi.RuneDelete {
		return en.handleBackspace(r)
	}

	// cancel
	if r == '\a' { // Ctrl-G
		return en.handleCancel(r)
	}

	// show list of candidates
	if en.conv.index >= candOffset && r != ' ' && r != 'x' {
		return en.handleCandList(r)
	}

	// toggle line buffer mode
	if r == '\f' { // Ctrl-L
		return en.handleLineMode(r)
	}
	en.linePass = false

	// meta
	if r == '\n' { // Ctrl-J
		return en.handleMeta(r)
	}

	// enter
	if r == termi.RuneEnter && en.conv.mode == convNone {
		return en.handleEnter(r)
	}

	// zenkaku mode
	if en.inputMode == inputZen {
		return en.handleZen(r)
	}

	// ASCII mode
	if en.inputMode == inputASCII {
		return en.handleRune(r)
	}

	//
	// now in Hiragana or Katakana mode
	//

	// assert
	if en.inputMode != inputHira && en.inputMode != inputKata {
		en.message = fmt.Sprintf(
			"assert: invalid inputMode == %d", en.inputMode,
		)
		return en.output(true)
	}

	// enter zenkaku mode
	if r == 'L' {
		return en.handleZenMode()
	}

	// enter in kana mode
	if r == termi.RuneEnter && en.conv.mode != convNone {
		return en.handleKanaEnter()
	}

	// escape
	if r == termi.RuneEscape {
		return en.handleEscape(r)
	}

	// control code
	if r < ' ' {
		return en.handleControlCode(r)
	}

	// conversion
	if r == ' ' && en.conv.mode != convNone {
		return en.handleConv()
	}

	// reverse conversion
	if r == 'x' && en.conv.hasCands() {
		return en.handleConvRev()
	}

	// enter abbrev mode
	if r == '/' {
		return en.enterAbbrevMode()
	}

	// handle abbrev
	if en.conv.mode == convAbbrev {
		if en.conv.hasCands() {
			en.flush()
			en.conv.reset()
			// fallthrough
		} else {
			en.conv.stem.WriteRune(r)
			return en.output(true)
		}
	}

	update := false
	if en.conv.hasCands() {
		tail, ok := en.conv.okuri.Tail()
		if en.conv.mode != convOkuri || ok && tail != 'っ' && tail != 'ッ' {
			s := en.conv.cand() + en.conv.okuri.String()
			if en.regMode {
				en.regBuf.WriteString(s)
			} else if en.lineMode {
				en.lineBuf.WriteString(s)
			} else {
				en.out.WriteString(s)
			}
			en.conv.reset()
			update = true
		}
	}

	// kigou
	kigou, ok := romaji.ToKigou[string(r)]
	if ok {
		return en.handleKigou(kigou, update)
	}

	// phase shift operation
	if r >= 'A' && r <= 'Z' {
		if en.inputBuf.String() == "n" {
			if en.inputMode == inputHira {
				en.conv.stem.WriteRune('ん')
			} else { // inputKata
				en.conv.stem.WriteRune('ン')
			}
			en.inputBuf.Reset()
		}

		en.conv.advanceMode()

		// to lower
		r += 'a' - 'A'
	}

	if r < 'a' || r > 'z' {
		return en.handleNonAlpha(r, update)
	} else if r == 'l' {
		return en.enterASCIIMode()
	} else if r == 'q' {
		return en.changeKanaType()
	} else {
		return en.handleAlpha(r, update)
	}
}
