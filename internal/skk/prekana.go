package skk

import (
	"fmt"

	"tea.kareha.org/cup/termi"

	"tea.kareha.org/cup/kakiko/internal/fep"
	"tea.kareha.org/cup/kakiko/internal/romaji"
)

const pleaseAnswer = "Please answer y or n."

func (en *Engine) handleDeleteCand(seq termi.Seq) (string, fep.Cmd) {
	if en.message != "" {
		en.message = ""
		return "", fep.CmdDraw
	}
	if seq.Kind != termi.SeqRune {
		en.message = pleaseAnswer
		return "", fep.CmdDraw
	}
	if seq.Rune == 'n' {
		en.deleteMode = false
		return "", fep.CmdDraw
	}
	if seq.Rune == 'y' {
		en.dics.Remove(en.conv.stem.String(), en.conv.cand())
		en.dics.RemoveDiff(en.conv.stem.String(), en.conv.cand())
		en.resetConv()

		en.deleteMode = false
		return "", fep.CmdDraw
	}
	en.message = pleaseAnswer
	return "", fep.CmdDraw
}

func (en *Engine) handleBackspace(r rune) (string, fep.Cmd) {
	if en.inputBuf.Len() > 0 {
		en.inputBuf.Reset()
		if en.conv.mode == convOkuri {
			en.conv.mode = convStem
			en.conv.stem.RemoveTail()
		}
		return en.output(true)
	}

	if en.conv.mode != convNone {
		if en.conv.hasCands() {
			en.dics.Add(en.conv.stem.String(), en.conv.cand())

			if en.conv.mode == convOkuri {
				en.conv.okuri.RemoveTail()
			} else {
				if en.regMode && en.conv.out.Len() < 1 {
					en.regBuf.WriteString(en.conv.cand())
					en.regBuf.RemoveTail()
				} else {
					en.conv.out.WriteString(en.conv.cand())
					en.conv.out.RemoveTail()
				}
				en.conv.clearCands()
				en.conv.stem.Reset()
			}
			en.flush()
			en.conv.reset() // do not use en.resetConv()
			return en.output(true)
		}

		if en.conv.okuri.RemoveTail() {
			return en.output(true)
		}

		if en.conv.stem.RemoveTail() {
			en.conv.clearCands()
			return en.output(true)
		}

		en.resetConv()
		return en.output(true)
	}

	if en.regMode {
		if en.regBuf.RemoveTail() {
			return en.output(true)
		}

		if en.conv.out.RemoveTail() {
			return en.output(true)
		}

		en.message = "Beginning of buffer"
		return en.output(true)
	}

	if en.lineMode && en.lineBuf.RemoveTail() {
		return en.output(true)
	}

	en.out.WriteRune(r)
	return en.output(false)
}

func (en *Engine) handleCancel(r rune) (string, fep.Cmd) {
	switch en.conv.mode {
	default: //case convNone:
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
			if en.conv.mode == convOkuri {
				en.conv.stem.RemoveTail()
				en.conv.stem.WriteString(en.conv.okuri.String())
				en.conv.okuri.Reset()
				en.conv.mode = convStem
			}
			return en.output(true)
		}

		if en.inputBuf.Len() > 0 {
			en.inputBuf.Reset()
			return en.output(true)
		}
		en.out.WriteRune(r)
		return en.output(false)
	case convOkuri:
		en.inputBuf.Reset()
		if en.conv.hasCands() {
			en.conv.stem.RemoveTail()
			en.conv.stem.WriteString(en.conv.okuri.String())
			en.conv.okuri.Reset()
			en.conv.clearCands()
			en.conv.mode = convStem
		} else {
			en.resetConv()
		}
		return en.output(true)
	case convStem, convAbbrev:
		en.inputBuf.Reset()
		if en.conv.hasCands() {
			en.conv.clearCands()
		} else {
			en.resetConv()
		}
		return en.output(true)
	}
}

func (en *Engine) handleCandList(r rune) (string, fep.Cmd) {
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
	en.conv.reset() // do not use en.resetConv()
	return en.output(true)
}

func (en *Engine) handleLineMode(r rune) (string, fep.Cmd) {
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

func (en *Engine) handleSync() (string, fep.Cmd) {
	en.message = "User dictionaries synchronized."
	en.lineMode = false
	en.linePass = false
	s, _ := en.output(true)
	return s, fep.CmdSync
}

func (en *Engine) handleSuper(r rune) (string, fep.Cmd) {
	if en.conv.mode == convAbbrev && !en.conv.hasCands() {
		en.out.WriteString(romaji.HanToZen(en.conv.stem.String()))
		en.resetConv()
		return en.output(true)
	}

	if en.conv.mode != convNone {
		en.inputBuf.Reset()
		en.flush()
		en.conv.reset() // do not use en.resetConv()
		return en.output(true)
	}

	if en.inputMode != inputHira && en.inputMode != inputKata {
		en.inputMode = inputHira
		return en.output(true)
	}

	en.out.WriteRune(r)
	return en.output(false)
}

func (en *Engine) handleEnter(r rune) (string, fep.Cmd) {
	if en.regMode {
		en.flush()
		en.endReg()
		en.conv.clearCands()

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
		if en.regMode {
			return en.output(true)
		}

		en.flushPartial()
		en.resetConv()
		return en.output(true)
	}

	if en.lineMode && en.lineBuf.Len() > 0 {
		en.out.WriteString(en.lineBuf.String())
		en.lineBuf.Reset()
		return en.output(true)
	}

	en.out.WriteRune(r)
	return en.output(false)
}

func (en *Engine) handleZen(r rune) (string, fep.Cmd) {
	zen, ok := romaji.ToZen[string(r)]
	if ok {
		en.writeString(zen)
	} else {
		en.writeRune(r)
	}
	return en.output(en.regMode || en.lineMode)
}

func (en *Engine) handleRune(r rune) (string, fep.Cmd) {
	en.writeRune(r)
	return en.output(en.regMode || en.lineMode)
}
