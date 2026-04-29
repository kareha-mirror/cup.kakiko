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

	// show list of candidates
	if en.conv.index >= candOffset {
		switch r {
		case '\a': // Ctrl-G
			return en.handleCancel(r)
		case ' ': // conversion
			return en.handleConv()
		case 'x': // reverse conversion
			return en.handleConvRev()
		default:
			return en.handleCandList(r)
		}
	}

	// toggle line buffer mode
	if r == '\f' { // Ctrl-L
		return en.handleLineMode(r)
	}
	en.linePass = false

	switch r {
	case termi.RuneBackspace, termi.RuneDelete:
		return en.handleBackspace(r)
	case '\a': // Ctrl-G
		return en.handleCancel(r)
	case '\n': // Ctrl-J
		return en.handleMeta(r)
	}

	// enter out of conv mode
	if r == termi.RuneEnter && en.conv.mode == convNone {
		return en.handleEnter(r)
	}

	switch en.inputMode {
	case inputASCII:
		return en.handleRune(r)
	case inputZen:
		return en.handleZen(r)
	}

	// assert: now in Hiragana or Katakana mode
	if en.inputMode != inputHira && en.inputMode != inputKata {
		en.message = fmt.Sprintf(
			"assert: invalid inputMode == %d", en.inputMode,
		)
		return en.output(true)
	}

	switch r {
	case termi.RuneEnter:
		return en.handleConvEnter()
	case termi.RuneEscape:
		return en.handleEscape(r)
	case 'L':
		return en.enterZenMode()
	case '/':
		return en.enterAbbrevMode()
	}

	// control code
	// place this after handling Enter, Escape or other control code
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
			en.write(en.conv.cand() + en.conv.okuri.String())
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
		// to lower
		r += 'a' - 'A'

		inp := en.inputBuf.String() + string(r)
		if _, ok := romaji.IsSokuon[inp]; ok {
			if en.inputMode == inputHira {
				en.conv.stem.WriteRune('っ')
			} else { // inputKata
				en.conv.stem.WriteRune('ッ')
			}
			en.inputBuf.RemoveHead()
		} else if _, ok := romaji.IsN[inp]; ok {
			if en.inputMode == inputHira {
				en.conv.stem.WriteRune('ん')
			} else { // inputKata
				en.conv.stem.WriteRune('ン')
			}
			en.inputBuf.RemoveHead()
		}

		en.conv.advanceMode()
	}

	if r < 'a' || r > 'z' {
		return en.handleNonAlpha(r, update)
	}

	switch r {
	case 'l':
		return en.enterASCIIMode()
	case 'q':
		return en.changeKanaType()
	}

	return en.handleAlpha(r, update)
}
