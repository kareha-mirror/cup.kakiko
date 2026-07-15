package skk

import "tea.kareha.org/cup/termi/rbuf"

type convMode int

const (
	convNone convMode = iota
	convStem
	convOkuri
	convAbbrev
)

type conv struct {
	mode convMode

	out   rbuf.RuneBuf
	stem  rbuf.RuneBuf
	okuri rbuf.RuneBuf
	tail  rbuf.RuneBuf

	cands []string
	index int
}

func newConv() *conv {
	return &conv{
		mode: convNone,

		out:   rbuf.RuneBuf{},
		stem:  rbuf.RuneBuf{},
		okuri: rbuf.RuneBuf{},
		tail:  rbuf.RuneBuf{},

		cands: []string{},
		index: 0,
	}
}

func (c *conv) clearCands() {
	c.cands = c.cands[:0]
	c.index = 0
}

func (c *conv) resetPartial() {
	c.mode = convNone

	c.stem.Reset()
	c.okuri.Reset()

	c.clearCands()
}

func (c *conv) reset() {
	c.resetPartial()
	c.tail.Reset()
	c.out.Reset()
}

func (c *conv) hasCands() bool {
	return len(c.cands) > 0
}

func (c *conv) candByIndex(index int) string {
	if !c.hasCands() {
		return ""
	}
	return c.cands[index]
}

func (c *conv) cand() string {
	return c.candByIndex(c.index)
}

const candOffset = 4

var (
	candKeyList = []rune{'a', 's', 'd', 'f', 'j', 'k', 'l'}
	candKeys    = map[rune]int{}
)

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
	} else if c.mode == convStem && c.stem.RuneCount() > 0 {
		c.mode = convOkuri
	}
}

func (c *conv) stemBody() string {
	if c.mode == convOkuri {
		stem := c.stem.Body(0, c.stem.RuneCount()-1)
		if stem == nil {
			return ""
		}
		return stem.String()
	} else {
		return c.stem.String()
	}
}
