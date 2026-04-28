package skk

import (
	"strings"

	"tea.kareha.org/cup/termi"
)

type convMode int

const (
	convNone convMode = iota
	convStem
	convOkuri
	convAbbrev
)

type convState struct {
	mode  convMode
	out   termi.RuneBuf
	stem  termi.RuneBuf
	okuri termi.RuneBuf
	cands []string
	index int
}

func newConvState() *convState {
	return &convState{
		mode:  convNone,
		out:   termi.RuneBuf{},
		stem:  termi.RuneBuf{},
		okuri: termi.RuneBuf{},
		cands: []string{},
		index: 0,
	}
}

func (conv *convState) clearCands() {
	conv.cands = conv.cands[:0]
	conv.index = 0
}

func (conv *convState) reset() {
	conv.mode = convNone
	conv.out.Reset()
	conv.stem.Reset()
	conv.okuri.Reset()
	conv.clearCands()
}

func (conv *convState) hasCands() bool {
	return len(conv.cands) > 0
}

func (conv *convState) candByIndex(index int) string {
	if !conv.hasCands() {
		return ""
	}

	cand := conv.cands[index]
	semicolon := strings.Index(cand, ";")
	if semicolon < 0 {
		return cand
	}
	return cand[:semicolon]
}

func (conv *convState) cand() string {
	return conv.candByIndex(conv.index)
}

const candOffset = 4

var candKeyList = []rune{'a', 's', 'd', 'f', 'j', 'k', 'l'}
var candKeys = map[rune]int{}

func init() {
	for i, r := range candKeyList {
		candKeys[r] = i
	}
}

func (conv *convState) keyToIndex(r rune) int {
	if conv.index < candOffset {
		return -1
	}
	i, ok := candKeys[r]
	if !ok {
		return -1
	}
	if conv.index+i >= len(conv.cands) {
		return -1
	}
	return conv.index + i
}

func (conv *convState) advanceModeOnUpper() {
	if conv.mode == convNone {
		conv.mode = convStem
	} else if conv.mode == convStem {
		conv.mode = convOkuri
	}
}
