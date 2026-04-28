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

type conv struct {
	mode convMode

	out   termi.RuneBuf
	stem  termi.RuneBuf
	okuri termi.RuneBuf

	cands []string
	index int
}

func newConv() *conv {
	return &conv{
		mode: convNone,

		out:   termi.RuneBuf{},
		stem:  termi.RuneBuf{},
		okuri: termi.RuneBuf{},

		cands: []string{},
		index: 0,
	}
}

func (c *conv) clearCands() {
	c.cands = c.cands[:0]
	c.index = 0
}

func (c *conv) reset() {
	c.mode = convNone

	c.out.Reset()
	c.stem.Reset()
	c.okuri.Reset()

	c.clearCands()
}

func (c *conv) hasCands() bool {
	return len(c.cands) > 0
}

func (c *conv) candByIndex(index int) string {
	if !c.hasCands() {
		return ""
	}

	cand := c.cands[index]
	semicolon := strings.Index(cand, ";")
	if semicolon < 0 {
		return cand
	}
	return cand[:semicolon]
}

func (c *conv) cand() string {
	return c.candByIndex(c.index)
}

const candOffset = 4

var candKeyList = []rune{'a', 's', 'd', 'f', 'j', 'k', 'l'}
var candKeys = map[rune]int{}

func init() {
	for i, r := range candKeyList {
		candKeys[r] = i
	}
}

func (c *conv) keyToIndex(r rune) int {
	if c.index < candOffset {
		return -1
	}
	i, ok := candKeys[r]
	if !ok {
		return -1
	}
	if c.index+i >= len(c.cands) {
		return -1
	}
	return c.index + i
}

func (c *conv) advanceMode() {
	if c.mode == convNone {
		c.mode = convStem
	} else if c.mode == convStem {
		c.mode = convOkuri
	}
}
