package skk

import (
	"fmt"
	"strings"

	"tea.kareha.org/cup/termi"

	"tea.kareha.org/cup/kakiko/internal/romaji"
)

type inputMode int

const (
	inputASCII inputMode = iota
	inputHira
	inputKata
	inputZen
)

type convMode int

const (
	convNone convMode = iota
	convGokan
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

	iprev int
	ccand string
}

func newConvState() *convState {
	return &convState{
		mode:  convNone,
		out:   termi.RuneBuf{},
		stem:  termi.RuneBuf{},
		okuri: termi.RuneBuf{},
		cands: []string{},
		index: 0,

		iprev: -1,
		ccand: "",
	}
}

func (conv *convState) clearCands() {
	conv.cands = conv.cands[:0]
	conv.index = 0
	conv.iprev = -1
	conv.ccand = ""
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
	if conv.index == conv.iprev {
		return conv.ccand
	}
	cand := conv.candByIndex(conv.index)
	if cand == "" {
		return cand
	}
	conv.iprev = conv.index
	conv.ccand = cand
	return conv.ccand
}

const candOffset = 4

var candKeys = []rune{'a', 's', 'd', 'f', 'j', 'k', 'l'}

func (conv *convState) keyToIndex(r rune) int {
	if conv.index < candOffset {
		return -1
	}
	for i := 0; i < len(candKeys); i++ {
		if conv.index+i >= len(conv.cands) {
			break
		}
		if r == candKeys[i] {
			return conv.index + i
		}
	}
	return -1
}

type Engine struct {
	d Dics

	inputMode inputMode
	inputBuf  termi.RuneBuf

	conv      *convState
	convStack []*convState
	regMode   bool
	regBuf    termi.RuneBuf

	lineMode bool
	lineBuf  termi.RuneBuf
	linePass bool

	message string
}

func NewEngine(path string) *Engine {
	d := Dics{}
	d.SetUserDic(NewMemUserDic())
	d.AddDic(NewCDBDic(path))
	en := &Engine{
		d: d,

		inputMode: inputASCII,
		inputBuf:  termi.RuneBuf{},
		conv:      newConvState(),
		convStack: []*convState{},
		regMode:   false,
		regBuf:    termi.RuneBuf{},

		lineMode: false,
		lineBuf:  termi.RuneBuf{},
		linePass: false,

		message: "",
	}

	return en
}

func (en *Engine) pushConv() {
	en.convStack = append(en.convStack, en.conv)
	en.conv = newConvState()
}

func (en *Engine) popConv() bool {
	n := len(en.convStack)
	if n < 1 {
		return false
	}
	en.conv = en.convStack[n-1]
	en.convStack = en.convStack[:n-1]
	return n > 1
}

var vowels = map[string]string{
	"あ": "a",
	"い": "i",
	"う": "u",
	"え": "e",
	"お": "o",
}

func vowelOf(kana string) (string, bool) {
	switch kana {
	case "あ":
		return "a", true
	case "い":
		return "i", true
	case "う":
		return "u", true
	case "え":
		return "e", true
	case "お":
		return "o", true
	default:
		return "", false
	}
}

func (en *Engine) Process(key termi.Key) (string, bool) {
	if en.message != "" {
		en.message = ""
		return "", true
	}

	if key.Kind != termi.KeyRune {
		return key.Raw, false
	}

	r := key.Rune

	if r == termi.RuneBackspace || r == termi.RuneDelete {
		if en.inputBuf.Len() > 0 {
			en.inputBuf.Reset()
			return "", true
		}
		if en.conv.okuri.RemoveTail() {
			return "", true
		}
		if en.conv.mode != convNone && en.conv.stem.RemoveTail() {
			if en.conv.stem.Len() < 1 {
				en.conv.mode = convNone
			}
			en.conv.clearCands()
			return "", true
		}
		if en.regMode {
			if en.regBuf.RemoveTail() {
				return "", true
			}
			if en.conv.out.RemoveTail() {
				return "", true
			}
			en.inputBuf.Reset()
			if !en.popConv() {
				en.regMode = false
			}
			return "", true
		}
		if en.lineMode && en.lineBuf.RemoveTail() {
			return "", true
		}
		return string(r), false
	}

	output := strings.Builder{}

	flush := func() {
		s := strings.Builder{}
		if en.conv.out.Len() > 0 {
			s.WriteString(en.conv.out.String())
		} else if en.conv.hasCands() {
			s.WriteString(en.conv.cand())
		} else {
			s.WriteString(en.conv.stem.String())
		}
		s.WriteString(en.conv.okuri.String())
		if en.regMode {
			en.regBuf.WriteString(s.String())
		} else if en.lineMode {
			en.lineBuf.WriteString(s.String())
		} else {
			output.WriteString(s.String())
		}
	}

	if r == '\a' { // Ctrl-G
		if en.regMode {
			if en.inputBuf.Len() > 0 {
				en.inputBuf.Reset()
				return output.String(), true
			}
			if en.regBuf.Len() > 0 || en.conv.out.Len() > 0 {
				en.regBuf.Reset()
				en.conv.out.Reset()
				return output.String(), true
			}
			if !en.popConv() {
				en.regMode = false
			}
			return output.String(), true
		}

		switch en.conv.mode {
		case convNone:
			if en.inputBuf.Len() > 0 {
				en.inputBuf.Reset()
				return output.String(), true
			}
			output.WriteRune(r)
			return output.String(), false
		case convOkuri:
			en.inputBuf.Reset()
			if !en.conv.hasCands() {
				en.conv.reset()
			} else {
				en.conv.stem.RemoveTail()
				en.conv.stem.WriteString(en.conv.okuri.String())
				en.conv.okuri.Reset()
				en.conv.clearCands()
				en.conv.mode = convGokan
			}
			return output.String(), true
		case convGokan, convAbbrev:
			en.inputBuf.Reset()
			if !en.conv.hasCands() {
				en.conv.reset()
			} else {
				en.conv.clearCands()
			}
			return output.String(), true
		}
	}

	if en.conv.index >= candOffset && r != ' ' && r != 'x' {
		index := en.conv.keyToIndex(r)
		if index < 0 {
			en.message = fmt.Sprintf(
				//"\"%c\" is not valid here!", r,
				"\"%c\" は有効なキーではありません！", r,
			)
			return output.String(), true
		}

		flush()
		en.conv.reset()
		return output.String(), true
	}

	if r == '\f' { // Ctrl-L
		if en.regMode {
			output.WriteRune(r)
			return output.String(), true
		}

		if en.linePass {
			en.linePass = false
			output.WriteRune(r)
		} else {
			en.linePass = true
		}

		if en.lineMode {
			en.lineMode = false
			flush()
			output.WriteString(en.lineBuf.String())
			en.lineBuf.Reset()
		} else {
			en.lineMode = true
		}

		return output.String(), true
	}
	en.linePass = false

	if r == '\n' { // Ctrl-J
		if en.conv.mode == convAbbrev {
			output.WriteString(romaji.HanToZen(en.conv.stem.String()))
			en.conv.reset()
			return output.String(), true
		}

		if en.conv.mode != convNone {
			flush()
		}

		if en.inputMode != inputHira && en.inputMode != inputKata {
			en.inputMode = inputHira
		}

		en.conv.reset()

		return output.String(), true
	}

	if r == termi.RuneEnter && en.conv.mode == convNone {
		reg := false
		if en.regMode {
			flush()
			en.inputBuf.Reset()
			if !en.popConv() {
				en.regMode = false
			}

			regWord := en.regBuf.String()
			en.regBuf.Reset()

			if en.conv.mode == convOkuri {
				en.d.AddOkuri(
					en.conv.stem.String(), en.conv.okuri.String(), regWord,
				)
			} else {
				en.d.Add(en.conv.stem.String(), regWord)
			}

			en.conv.out.WriteString(regWord)
			en.conv.stem.Reset()
			en.conv.okuri.Reset()
			en.conv.mode = convNone
			if en.regMode {
				return output.String(), true
			}

			flush()
			en.conv.reset()
			reg = true
		}

		if en.lineMode {
			if en.lineBuf.Len() > 0 {
				output.WriteString(en.lineBuf.String())
				en.lineBuf.Reset()
			} else {
				output.WriteRune(r)
			}
			return output.String(), true
		} else {
			if !reg {
				output.WriteRune(r)
			}
			return output.String(), reg
		}
	}

	if en.inputMode == inputZen {
		zen, ok := romaji.ToZen[string(r)]
		if ok {
			if en.regMode {
				en.regBuf.WriteString(zen)
				return output.String(), true
			} else if en.lineMode {
				en.lineBuf.WriteString(zen)
				return output.String(), true
			} else {
				output.WriteString(zen)
				return output.String(), false
			}
		} else {
			if en.regMode {
				en.regBuf.WriteRune(r)
				return output.String(), true
			} else if en.lineMode {
				en.lineBuf.WriteRune(r)
				return output.String(), true
			} else {
				output.WriteRune(r)
				return output.String(), false
			}
		}
	}

	if en.inputMode == inputASCII {
		if en.regMode {
			en.regBuf.WriteRune(r)
			return output.String(), true
		} else if en.lineMode {
			en.lineBuf.WriteRune(r)
			return output.String(), true
		} else {
			output.WriteRune(r)
			return output.String(), false
		}
	}

	// now in Hiragana or Katakana mode
	// assert
	if en.inputMode != inputHira && en.inputMode != inputKata {
		en.message = "invalid inputMode == " + string(en.inputMode)
		return output.String(), true
	}

	if r == 'L' {
		en.inputMode = inputZen
		en.inputBuf.Reset()

		if en.conv.mode != convNone {
			flush()
			en.conv.reset()
		}

		return output.String(), true
	}

	if r == termi.RuneEnter && en.conv.mode != convNone {
		if en.regMode {
			flush()
			en.inputBuf.Reset()
			if !en.popConv() {
				en.regMode = false
			}

			regWord := en.regBuf.String()
			en.regBuf.Reset()

			if en.conv.mode == convOkuri {
				en.d.AddOkuri(
					en.conv.stem.String(), en.conv.okuri.String(), regWord,
				)
			} else {
				en.d.Add(en.conv.stem.String(), regWord)
			}

			en.conv.out.WriteString(regWord)
			en.conv.stem.Reset()
			en.conv.okuri.Reset()
			en.conv.mode = convNone
			if en.regMode {
				return output.String(), true
			}
		}

		flush()
		en.conv.reset()
		return output.String(), true
	}

	if r == termi.RuneEscape {
		en.inputMode = inputASCII
		en.inputBuf.Reset()

		if en.conv.mode != convNone {
			flush()
			en.conv.reset()
		}

		// XXX regMode?

		if en.lineMode {
			output.WriteString(en.lineBuf.String())
			en.lineBuf.Reset()
			en.lineMode = false
		}

		output.WriteRune(r)
		return output.String(), true
	}

	if r < ' ' {
		output.WriteRune(r)
		return output.String(), false
	}

	if r == ' ' && en.conv.mode != convNone {
		if !en.conv.hasCands() {
			if en.inputBuf.String() == "n" {
				if en.inputMode == inputHira {
					en.conv.stem.WriteString("ん")
				} else { // inputKata
					en.conv.stem.WriteString("ン")
				}
			}
			en.inputBuf.Reset()

			stem := en.conv.stem.String()
			stem = romaji.KataToHira(stem)
			var err error
			if en.conv.mode == convOkuri {
				okuri := en.conv.okuri.String()
				okuri = romaji.KataToHira(okuri)
				en.conv.cands, err = en.d.LookupOkuri(stem, okuri)
			} else {
				en.conv.cands, err = en.d.Lookup(stem)
			}
			en.conv.index = 0
			if err != nil {
				en.message = fmt.Sprintf("%v", err)
				en.conv.cands = []string{}
			} else if !en.conv.hasCands() {
				en.regMode = true
				en.pushConv()
				return output.String(), true
			}
		} else {
			if en.conv.index < candOffset {
				if en.conv.index+1 < len(en.conv.cands) {
					en.conv.index++
				} else {
					en.regMode = true
					en.pushConv()
					return output.String(), true
				}
			} else {
				if en.conv.index+len(candKeys) < len(en.conv.cands) {
					en.conv.index += len(candKeys)
				} else {
					en.regMode = true
					en.pushConv()
					return output.String(), true
				}
			}
		}
		return output.String(), true
	}

	if r == 'x' && en.conv.hasCands() {
		if en.conv.index > candOffset {
			en.conv.index -= len(candKeys)
		} else {
			en.conv.index--
		}
		if en.conv.index < 0 {
			en.conv.okuri.Reset()
			en.conv.clearCands()
		}
		return output.String(), true
	}

	if r == '/' {
		if en.conv.hasCands() {
			flush()
			en.conv.reset()
		}
		en.conv.mode = convAbbrev
		return output.String(), true
	}

	if en.conv.mode == convAbbrev {
		if en.conv.hasCands() {
			flush()
			en.conv.reset()
		} else {
			en.conv.stem.WriteRune(r)
			return output.String(), true
		}
	}

	update := false
	if en.conv.hasCands() {
		if en.conv.mode == convOkuri {
			okuri := en.conv.okuri.String()
			if en.conv.okuri.Len() > 0 &&
				!strings.HasSuffix(okuri, "っ") &&
				!strings.HasSuffix(okuri, "ッ") {
				if en.regMode {
					en.regBuf.WriteString(en.conv.cand())
					en.regBuf.WriteString(okuri)
				} else if en.lineMode {
					en.lineBuf.WriteString(en.conv.cand())
					en.lineBuf.WriteString(okuri)
				} else {
					output.WriteString(en.conv.cand())
					output.WriteString(okuri)
				}
				en.conv.reset()
				en.inputBuf.Reset()
				update = true
			}
		} else {
			if en.regMode {
				en.regBuf.WriteString(en.conv.cand())
				en.regBuf.WriteString(en.conv.okuri.String())
			} else if en.lineMode {
				en.lineBuf.WriteString(en.conv.cand())
				en.lineBuf.WriteString(en.conv.okuri.String())
			} else {
				output.WriteString(en.conv.cand())
				output.WriteString(en.conv.okuri.String())
			}
			en.conv.reset()
			update = true
		}
	}

	if r >= 'A' && r <= 'Z' {
		if en.inputBuf.String() == "n" {
			if en.inputMode == inputHira {
				en.conv.stem.WriteString("ん")
			} else { // inputKata
				en.conv.stem.WriteString("ン")
			}
			en.inputBuf.Reset()
		}
		if en.conv.mode == convNone {
			en.conv.mode = convGokan
		} else if en.conv.mode == convGokan {
			en.conv.mode = convOkuri
		}
		r += 'a' - 'A'
	}

	kigou, ok := romaji.ToKigou[string(r)]
	if ok {
		if en.conv.mode != convNone && kigou == "ー" {
			if !en.conv.hasCands() {
				en.conv.stem.WriteString(kigou)
				return output.String(), true
			}
		}

		update = update || en.inputBuf.Len() > 0
		en.inputBuf.Reset()
		if en.regMode {
			en.regBuf.WriteString(kigou)
			update = true
		} else if en.lineMode {
			en.lineBuf.WriteString(kigou)
			update = true
		} else {
			output.WriteString(kigou)
		}
		return output.String(), update
	}

	if r < 'a' || r > 'z' {
		if en.conv.mode != convNone {
			flush()
			en.conv.reset()
			update = true
		}

		if en.regMode {
			en.regBuf.WriteRune(r)
			update = true
		} else if en.lineMode {
			en.lineBuf.WriteRune(r)
			update = true
		} else {
			output.WriteRune(r)
		}
		return output.String(), update
	}

	if r == 'l' {
		en.inputMode = inputASCII
		en.inputBuf.Reset()

		if en.conv.mode != convNone {
			flush()
			en.conv.reset()
		}

		return output.String(), true
	}

	if r == 'q' {
		if en.conv.mode == convNone {
			if en.inputMode == inputHira {
				en.inputMode = inputKata
			} else { // inputKata
				en.inputMode = inputHira
			}
			return output.String(), true
		} else if en.conv.mode == convGokan {
			if en.inputBuf.String() == "n" {
				if en.inputMode == inputHira {
					en.conv.stem.WriteString("ん")
				} else { // inputKata
					en.conv.stem.WriteString("ン")
				}
			}
			en.inputBuf.Reset()

			en.conv.mode = convNone
			var s string
			if en.inputMode == inputHira {
				s = romaji.HiraToKata(en.conv.stem.String())
			} else { // inputKata
				s = romaji.KataToHira(en.conv.stem.String())
			}
			en.conv.stem.Reset()
			if en.regMode {
				en.regBuf.WriteString(s)
			} else if en.lineMode {
				en.lineBuf.WriteString(s)
			} else {
				output.WriteString(s)
			}
			return output.String(), true
		}
	}

	en.inputBuf.WriteRune(r)

	var kana string
	sokuon := false
	if _, ok := romaji.IsSokuon[en.inputBuf.String()]; ok {
		if en.inputMode == inputHira {
			kana = "っ"
		} else { // inputKata
			kana = "ッ"
		}
		en.inputBuf.RemoveHead()
		sokuon = true
	} else if _, ok := romaji.IsN[en.inputBuf.String()]; ok {
		if en.inputMode == inputHira {
			kana = "ん"
		} else { // inputKata
			kana = "ン"
		}
		en.inputBuf.RemoveHead()
	} else {
		lookup := en.inputBuf.String()
		alias, ok := romaji.Aliases[lookup]
		if ok {
			lookup = alias
		}

		var k string
		if en.inputMode == inputHira {
			k, ok = romaji.ToHira[lookup]
		} else { // inputKata
			k, ok = romaji.ToKata[lookup]
		}
		if ok {
			kana = k
			en.inputBuf.Reset()
		}
	}

	if en.conv.mode == convNone {
		if kana != "" {
			if en.regMode {
				en.regBuf.WriteString(kana)
			} else if en.lineMode {
				en.lineBuf.WriteString(kana)
			} else {
				output.WriteString(kana)
			}
		}
		return output.String(), true
	} else if en.conv.mode == convGokan {
		if kana != "" {
			en.conv.stem.WriteString(kana)
			en.conv.clearCands()
		}
		return output.String(), true
	} else if en.conv.mode == convOkuri {
		vowel, ok := vowelOf(kana)
		if ok {
			en.conv.stem.WriteString(vowel)
			en.conv.okuri.WriteString(kana)
		} else if kana != "" {
			en.conv.okuri.WriteString(kana)
		} else {
			en.conv.stem.WriteRune(r)
		}

		if en.conv.okuri.Len() < 1 {
			return output.String(), true
		}

		if sokuon {
			return output.String(), true
		}

		stem := en.conv.stem.String()
		stem = romaji.KataToHira(stem)
		okuri := en.conv.okuri.String()
		okuri = romaji.KataToHira(okuri)
		var err error
		en.conv.cands, err = en.d.LookupOkuri(stem, okuri)
		en.conv.index = 0
		if err != nil {
			en.message = fmt.Sprintf("%v", err)
		}
		return output.String(), true
	} else {
		en.message = "Process: invalid conv.mode == " + string(en.conv.mode)
		return output.String(), false
	}
}

func (en *Engine) Status() (string, bool) {
	if en.message != "" {
		return en.message, true
	}

	if en.conv.index >= candOffset {
		s := strings.Builder{}

		for i := 0; i < len(candKeys); i++ {
			if en.conv.index+i >= len(en.conv.cands) {
				break
			}

			s.WriteRune(candKeys[i] + 'A' - 'a')
			s.WriteRune(':')

			cand := en.conv.candByIndex(en.conv.index + i)
			s.WriteString(cand)
			s.WriteString("  ")
		}

		k := max(len(en.conv.cands)-en.conv.index-len(candKeys), 0)
		s.WriteString(fmt.Sprintf("[残り %d]", k))

		return s.String(), false
	}

	s := strings.Builder{}
	s.WriteRune('(')
	switch en.inputMode {
	case inputASCII:
		s.WriteString("SKK")
	case inputHira:
		s.WriteString("かな")
	case inputKata:
		s.WriteString("カナ")
	case inputZen:
		s.WriteString("全英")
	default:
		s.WriteString("不明")
	}
	s.WriteRune(')')

	if en.lineMode {
		s.WriteString(en.lineBuf.String())
		s.WriteRune(':')
	}

	if en.regMode {
		s.WriteString("[登録]")
		conv := en.convStack[len(en.convStack)-1]
		s.WriteString(conv.stem.String())
		s.WriteRune(' ')
		s.WriteString(en.conv.out.String())
		s.WriteString(en.regBuf.String())
	}

	if en.conv.mode != convNone {
		if en.conv.hasCands() {
			s.WriteRune('▼')
			s.WriteString(en.conv.cand())
		} else {
			s.WriteRune('▽')
			var stem string
			if en.conv.mode == convOkuri {
				if en.conv.stem.Len() < 1 {
					stem = ""
				} else {
					stem = en.conv.stem.Substring(0, en.conv.stem.Len()-1)
				}
			} else {
				stem = en.conv.stem.String()
			}
			s.WriteString(stem)
		}
	}

	if en.conv.mode == convOkuri && en.inputBuf.Len() > 0 {
		s.WriteRune('*')
	}
	s.WriteString(en.conv.okuri.String())
	s.WriteString(en.inputBuf.String())

	s.WriteRune('_')

	return s.String(), false
}
