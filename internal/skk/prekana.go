package skk

import (
	"fmt"

	"tea.kareha.org/cup/kakiko/internal/romaji"
)

func (en *Engine) handleBackspace(r rune) (string, bool) {
	if en.inputBuf.Len() > 0 {
		en.inputBuf.Reset()
		return en.output(true)
	}

	if en.conv.okuri.RemoveTail() {
		return en.output(true)
	}

	if en.conv.mode != convNone {
		if en.conv.hasCands() {
			en.conv.out.WriteString(en.conv.cand())
			en.conv.out.RemoveTail()
			en.flush()
			en.conv.reset()
			return en.output(true)
		}

		if en.conv.stem.RemoveTail() {
			if en.conv.stem.Len() < 1 {
				en.conv.mode = convNone
			}
			en.conv.clearCands()
			return en.output(true)
		}
	}

	if en.regMode {
		if en.regBuf.RemoveTail() {
			return en.output(true)
		}

		if en.conv.out.RemoveTail() {
			return en.output(true)
		}

		en.endReg()
		return en.output(true)
	}

	if en.lineMode && en.lineBuf.RemoveTail() {
		return en.output(true)
	}

	en.out.WriteRune(r)
	return en.output(false)
}

func (en *Engine) handleCancel(r rune) (string, bool) {
	if en.regMode {
		if en.inputBuf.Len() > 0 {
			en.inputBuf.Reset()
			return en.output(true)
		}

		if en.regBuf.Len() > 0 || en.conv.out.Len() > 0 {
			en.regBuf.Reset()
			en.conv.out.Reset()
			return en.output(true)
		}

		en.endReg()
		return en.output(true)
	}

	switch en.conv.mode {
	default: //case convNone:
		if en.inputBuf.Len() > 0 {
			en.inputBuf.Reset()
			return en.output(true)
		}
		en.out.WriteRune(r)
		return en.output(false)
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
		return en.output(true)
	case convStem, convAbbrev:
		en.inputBuf.Reset()
		if !en.conv.hasCands() {
			en.conv.reset()
		} else {
			en.conv.clearCands()
		}
		return en.output(true)
	}
}

func (en *Engine) handleCandList(r rune) (string, bool) {
	index := en.conv.keyToIndex(r)
	if index < 0 {
		en.message = fmt.Sprintf(
			//"\"%c\" is not valid here!", r,
			"\"%c\" は有効なキーではありません！", r,
		)
		return en.output(true)
	}
	en.conv.index = index

	en.flush()
	en.conv.reset()
	return en.output(true)
}

func (en *Engine) handleLineMode(r rune) (string, bool) {
	if en.regMode {
		en.out.WriteRune(r)
		return en.output(true)
	}

	if en.linePass {
		en.linePass = false
		en.out.WriteRune(r)
	} else {
		en.linePass = true
	}

	if en.lineMode {
		en.lineMode = false
		en.flush()
		en.out.WriteString(en.lineBuf.String())
		en.lineBuf.Reset()
	} else {
		en.lineMode = true
	}

	return en.output(true)
}

func (en *Engine) handleMeta(r rune) (string, bool) {
	if en.conv.mode == convAbbrev && !en.conv.hasCands() {
		en.out.WriteString(romaji.HanToZen(en.conv.stem.String()))
		en.conv.reset()
		return en.output(true)
	}

	if en.conv.mode != convNone {
		en.flush()
	}

	if en.inputMode != inputHira && en.inputMode != inputKata {
		en.inputMode = inputHira
	}

	en.conv.reset()

	return en.output(true)
}

func (en *Engine) handleEnter(r rune) (string, bool) {
	reg := false
	if en.regMode {
		en.flush()
		en.endReg()

		kanji := en.regBuf.String()
		en.regBuf.Reset()

		if en.conv.mode == convOkuri {
			en.dics.AddOkuri(
				en.conv.stem.String(), en.conv.okuri.String(), kanji,
			)
		} else {
			en.dics.Add(en.conv.stem.String(), kanji)
		}

		en.conv.out.WriteString(kanji)
		en.conv.stem.Reset()
		en.conv.okuri.Reset()
		en.conv.mode = convNone
		if en.regMode {
			return en.output(true)
		}

		en.flush()
		en.conv.reset()
		reg = true
	}

	if en.lineMode && en.lineBuf.Len() > 0 {
		en.out.WriteString(en.lineBuf.String())
		en.lineBuf.Reset()
		return en.output(true)
	}

	if !reg {
		en.out.WriteRune(r)
	}
	return en.output(reg)
}

func (en *Engine) handleZen(r rune) (string, bool) {
	zen, ok := romaji.ToZen[string(r)]
	if ok {
		if en.regMode {
			en.regBuf.WriteString(zen)
			return en.output(true)
		} else if en.lineMode {
			en.lineBuf.WriteString(zen)
			return en.output(true)
		} else {
			en.out.WriteString(zen)
			return en.output(false)
		}
	} else {
		if en.regMode {
			en.regBuf.WriteRune(r)
			return en.output(true)
		} else if en.lineMode {
			en.lineBuf.WriteRune(r)
			return en.output(true)
		} else {
			en.out.WriteRune(r)
			return en.output(false)
		}
	}
}

func (en *Engine) handleRune(r rune) (string, bool) {
	if en.regMode {
		en.regBuf.WriteRune(r)
		return en.output(true)
	} else if en.lineMode {
		en.lineBuf.WriteRune(r)
		return en.output(true)
	} else {
		en.out.WriteRune(r)
		return en.output(false)
	}
}
