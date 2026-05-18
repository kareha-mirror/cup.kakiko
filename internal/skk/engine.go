package skk

import (
	"strings"

	"tea.kareha.org/cup/termi"

	"tea.kareha.org/cup/kakiko/internal/skkdic"
)

type inputMode int

const (
	inputASCII inputMode = iota
	inputHira
	inputKata
	inputZen
)

type Engine struct {
	dics skkdic.Dics

	inputMode inputMode
	inputBuf  termi.RuneBuf
	conv      *conv

	stack   []*conv
	regMode bool
	regBuf  termi.RuneBuf

	lineMode bool
	lineBuf  termi.RuneBuf
	linePass bool

	pasteMode bool

	message string

	out strings.Builder
}

func NewEngine(builtinDic, path, userPath, diffPath string) *Engine {
	dics := skkdic.Dics{}
	dics.AddDic(skkdic.NewStrDic(builtinDic))
	dics.AddDic(skkdic.NewCDBDic(path))
	dics.SetUserDic(skkdic.NewMemDic(userPath))
	dics.SetDiffDic(skkdic.NewMemDic(diffPath))

	en := &Engine{
		dics: dics,

		inputMode: inputASCII,
		inputBuf:  termi.RuneBuf{},
		conv:      newConv(),

		stack:   []*conv{},
		regMode: false,
		regBuf:  termi.RuneBuf{},

		lineMode: false,
		lineBuf:  termi.RuneBuf{},
		linePass: false,

		pasteMode: false,

		message: "",

		out: strings.Builder{},
	}

	return en
}

func (en *Engine) Init() error {
	return nil
}

func (en *Engine) Finish() error {
	return en.dics.Finish()
}

func (en *Engine) output(update bool) (string, bool) {
	s := en.out.String()
	en.out.Reset()
	return s, update
}

func (en *Engine) pushConv() {
	en.stack = append(en.stack, en.conv)
	en.conv = newConv()
}

func (en *Engine) popConv() bool {
	n := len(en.stack)
	if n < 1 {
		return false
	}
	en.conv = en.stack[n-1]
	en.stack = en.stack[:n-1]
	return n > 1
}

func (en *Engine) beginReg() {
	en.regMode = true
	en.conv.out.WriteString(en.regBuf.String())
	en.regBuf.Reset()
	en.pushConv()
}

func (en *Engine) endReg() {
	en.inputBuf.Reset()
	if !en.popConv() {
		en.regMode = false
	}
}

func (en *Engine) write(s string) {
	if en.regMode {
		en.regBuf.WriteString(s)
	} else if en.lineMode {
		en.lineBuf.WriteString(s)
	} else {
		en.out.WriteString(s)
	}
}

func (en *Engine) flush() {
	s := strings.Builder{}
	s.WriteString(en.conv.out.String())

	if en.conv.hasCands() && en.conv.stem.Len() > 0 {
		s.WriteString(en.conv.cand())
		en.dics.Add(en.conv.stem.String(), en.conv.cand())
	} else {
		s.WriteString(en.conv.trueStem())
	}

	s.WriteString(en.conv.okuri.String())
	en.write(s.String())
}

func (en *Engine) flushReg() {
	s := strings.Builder{}
	s.WriteString(en.conv.out.String())

	s.WriteString(en.conv.okuri.String())
	en.write(s.String())
}
