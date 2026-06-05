package skk

import (
	"path/filepath"
	"strings"

	"tea.kareha.org/cup/termi"

	"tea.kareha.org/cup/kakiko/internal/fep"
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
	save bool

	inputMode inputMode
	inputBuf  termi.RuneBuf
	conv      *conv

	stack   []*conv
	regMode bool
	regBuf  termi.RuneBuf

	lineMode bool
	lineBuf  termi.RuneBuf
	linePass bool

	deleteMode bool

	pasteMode bool

	message string

	out strings.Builder
}

func NewEngine(dics []string) *Engine {
	d := skkdic.Dics{}
	for _, dic := range dics {
		if dic == "" {
			continue
		}
		d.AddDic(skkdic.NewStrDic(dic))
	}

	return &Engine{
		dics: d,
		save: true,

		inputMode: inputASCII,
		inputBuf:  termi.RuneBuf{},
		conv:      newConv(),

		stack:   []*conv{},
		regMode: false,
		regBuf:  termi.RuneBuf{},

		lineMode: false,
		lineBuf:  termi.RuneBuf{},
		linePass: false,

		deleteMode: false,

		pasteMode: false,

		message: "",

		out: strings.Builder{},
	}
}

func (en *Engine) AddDic(dic skkdic.Dic) {
	en.dics.AddDic(dic)
}

func (en *Engine) SetUserDic(dic skkdic.UserDic) {
	en.dics.SetUserDic(dic)
}

func (en *Engine) SetDiffDic(dic skkdic.UserDic) {
	en.dics.SetDiffDic(dic)
}

func getSKKDicPath(dir string) string {
	return filepath.Join(dir, "skk-edic-legacy-large.cdb")
}

func getSKKUserDicPath(dir string) string {
	return filepath.Join(dir, "skk-edic-user.txt")
}

func getSKKDiffDicPath(dir string) string {
	return filepath.Join(dir, "skk-edic-diff.txt")
}

func (en *Engine) Init(dir string) error {
	dicPath := getSKKDicPath(dir)
	mainDic := skkdic.NewCDBDic(dicPath)
	en.AddDic(mainDic)

	userDicPath := getSKKUserDicPath(dir)
	userDic := skkdic.NewMemDic(userDicPath)
	en.SetUserDic(userDic)

	diffDicPath := getSKKDiffDicPath(dir)
	diffDic := skkdic.NewMemDic(diffDicPath)
	en.SetDiffDic(diffDic)

	return nil
}

func (en *Engine) Finish() error {
	return en.dics.Finish(en.save)
}

func (en *Engine) Sync() error {
	return en.dics.Sync()
}

func (en *Engine) output(update bool) (string, fep.Cmd) {
	s := en.out.String()
	en.out.Reset()
	if update {
		return s, fep.CmdDraw
	} else {
		return s, fep.CmdNone
	}
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

func (en *Engine) writeRune(r rune) {
	if en.regMode {
		en.regBuf.WriteRune(r)
	} else if en.lineMode {
		en.lineBuf.WriteRune(r)
	} else {
		en.out.WriteRune(r)
	}
}

func (en *Engine) writeString(s string) {
	if en.regMode {
		en.regBuf.WriteString(s)
	} else if en.lineMode {
		en.lineBuf.WriteString(s)
	} else {
		en.out.WriteString(s)
	}
}

func (en *Engine) flushPartial() {
	s := strings.Builder{}
	s.WriteString(en.conv.out.String())

	s.WriteString(en.conv.okuri.String())
	s.WriteString(en.conv.tail.String())
	en.writeString(s.String())
}

func (en *Engine) flush() {
	s := strings.Builder{}
	s.WriteString(en.conv.out.String())

	if en.conv.hasCands() {
		if en.conv.stem.Len() > 0 {
			s.WriteString(en.conv.cand())
			en.dics.Add(en.conv.stem.String(), en.conv.cand())
		}
	} else {
		s.WriteString(en.conv.stemBody())
	}

	s.WriteString(en.conv.okuri.String())
	s.WriteString(en.conv.tail.String())
	en.writeString(s.String())
}

func (en *Engine) resetConv() {
	if en.regMode {
		en.conv.resetPartial()
	} else {
		en.conv.reset()
	}
}
