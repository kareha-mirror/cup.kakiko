package skk

import (
	"tea.kareha.org/cup/termi"

	"tea.kareha.org/cup/kakiko/internal/fep"
	"tea.kareha.org/cup/kakiko/internal/romaji"
)

func (en *Engine) Process(key termi.Key) (string, fep.Cmd) {
	// delete candidate
	if en.deleteMode {
		return en.handleDeleteCand(key)
	}

	// paste
	if en.pasteMode {
		if key.Kind == termi.KeyEndPaste {
			en.pasteMode = false
			return "", fep.CmdNone
		}
		if key.Kind != termi.KeyRune {
			return "", fep.CmdNone
		}
		return en.handleRune(key.Rune)
	}
	if key.Kind == termi.KeyBeginPaste && en.regMode {
		en.pasteMode = true
		return "", fep.CmdNone
	}

	// hide message
	if en.message != "" {
		en.message = ""
		return "", fep.CmdDraw
	}

	// pass-through
	if key.Kind != termi.KeyRune {
		return key.Raw, fep.CmdNone
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
	if r == '\a' && en.linePass { // Ctrl-G
		return en.handleSync()
	}
	if r == '\x11' && en.linePass { // Ctrl-Q
		return en.toggleSave()
	}
	en.linePass = false

	// special keys: Backspace, Ctrl-G, Ctrl-J
	switch r {
	case termi.RuneBackspace, termi.RuneDelete:
		return en.handleBackspace(r)
	case '\a': // Ctrl-G
		return en.handleCancel(r)
	case termi.RuneNewline: // Ctrl-J
		return en.handleSuper(r)
	}

	// enter in non-conv mode
	if r == termi.RuneEnter && en.conv.mode == convNone {
		return en.handleEnter(r)
	}

	// Han to Zen - Ctrl-Q
	if r == '\x11' && en.conv.mode == convAbbrev && !en.conv.hasCands() {
		return en.handleToZen()
	}

	switch en.inputMode {
	case inputASCII:
		return en.handleRune(r)
	case inputZen:
		return en.handleZen(r)
	}
	// now in Hira or Kata mode

	switch r {
	case termi.RuneEnter:
		return en.handleConvEnter()
	case termi.RuneEscape:
		return en.handleEscape(r)
	case 'L':
		return en.enterZenMode()
	case '/':
		return en.enterAbbrevMode()
	case 'Q':
		return en.enterConvMode()
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

	// delete candidate
	if r == 'X' && en.conv.hasCands() {
		en.deleteMode = true
		return "", fep.CmdDraw
	}

	// prefix
	if r == '>' && en.conv.mode == convStem && !en.conv.hasCands() {
		en.resetKanaInput()
		en.conv.stem.WriteRune(r)
		return en.handleConv()
	}
	// suffix
	if r == '>' && en.conv.hasCands() {
		s, cmd := en.handleConvEnter()
		en.conv.stem.WriteRune(r)
		en.conv.mode = convStem
		return s, cmd
	}

	// handle abbrev
	if en.conv.mode == convAbbrev {
		if en.conv.hasCands() {
			en.flush()
			en.resetConv()
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
			en.dics.Add(en.conv.stem.String(), en.conv.cand())

			en.writeString(en.conv.cand())
			en.writeString(en.conv.okuri.String())
			en.writeString(en.conv.tail.String())
			en.resetConv()
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
			en.removeHeadInputBuf()
		} else if _, ok := romaji.IsN[inp]; ok {
			if en.inputMode == inputHira {
				en.conv.stem.WriteRune('ん')
			} else { // inputKata
				en.conv.stem.WriteRune('ン')
			}
			en.removeHeadInputBuf()
		}

		en.conv.advanceMode()
	}

	if r == '\'' && en.inputBuf.String() == "n" {
		return en.handleAlphabet('n', update)
	}

	if r < 'a' || r > 'z' {
		return en.handleNonAlphabet(r)
	}

	switch r {
	case 'l':
		return en.enterASCIIMode()
	case 'q':
		return en.toggleKanaType()
	}

	return en.handleAlphabet(r, update)
}
