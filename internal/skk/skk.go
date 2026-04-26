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

type Engine struct {
	j Jisyo

	inputMode   inputMode
	inputBuf    *termi.RuneBuf
	convMode    convMode
	convBuf     *termi.RuneBuf
	hasCandList bool
	candList    []string
	candIndex   int
	cand        string
	okuriBuf    *termi.RuneBuf

	lineMode bool
	lineBuf  *termi.RuneBuf
	linePass bool

	message string
}

func NewEngine(path string) *Engine {
	en := &Engine{
		j: NewCDBJisyo(path),

		inputMode:   inputASCII,
		inputBuf:    new(termi.RuneBuf),
		convMode:    convNone,
		convBuf:     new(termi.RuneBuf),
		hasCandList: false,
		candList:    []string{},
		candIndex:   0,
		cand:        "",
		okuriBuf:    new(termi.RuneBuf),

		lineMode: false,
		lineBuf:  new(termi.RuneBuf),
		linePass: false,

		message: "",
	}

	return en
}

const candOffset = 4

var candKeys = []rune{'a', 's', 'd', 'f', 'j', 'k', 'l'}

func (en *Engine) keyToCandIndex(r rune) int {
	if en.candIndex < candOffset {
		return -1
	}
	for i := 0; i < len(candKeys); i++ {
		if en.candIndex+i >= len(en.candList) {
			break
		}
		if r == candKeys[i] {
			return en.candIndex + i
		}
	}
	return -1
}

func (en *Engine) candByIndex(index int) string {
	if index < 0 || index >= len(en.candList) { // guard
		return "(Program Error)"
	}
	cand := en.candList[index]
	semicolon := strings.Index(cand, ";")
	if semicolon < 0 {
		return cand
	}
	return cand[:semicolon]
}

var vowels = map[string]string{
	"あ": "a",
	"い": "i",
	"う": "u",
	"え": "e",
	"お": "o",
}

func (en *Engine) resetConv() {
	en.convMode = convNone
	en.convBuf.Reset()
	en.hasCandList = false
	en.candList = []string{}
	en.candIndex = 0
	en.cand = ""
	en.okuriBuf.Reset()
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
		if en.okuriBuf.RemoveTail() {
			return "", true
		}
		if en.convMode != convNone && en.convBuf.RemoveTail() {
			if en.convBuf.Len() < 1 {
				en.convMode = convNone
			}
			en.hasCandList = false
			en.candList = []string{}
			en.candIndex = 0
			en.cand = ""
			return "", true
		}
		if en.lineMode && en.lineBuf.RemoveTail() {
			return "", true
		}
		return string(r), false
	}

	output := new(strings.Builder)

	flush := func() {
		s := new(strings.Builder)
		if en.cand != "" {
			s.WriteString(en.cand)
		} else {
			s.WriteString(en.convBuf.String())
		}
		s.WriteString(en.okuriBuf.String())
		if en.lineMode {
			en.lineBuf.WriteString(s.String())
		} else {
			output.WriteString(s.String())
		}
	}

	if r == '\a' { // Ctrl-G
		switch en.convMode {
		case convNone:
			if en.inputBuf.Len() > 0 {
				en.inputBuf.Reset()
				return output.String(), true
			}
			// not return
		case convOkuri:
			en.inputBuf.Reset()
			en.convBuf.RemoveTail()
			en.convBuf.WriteString(en.okuriBuf.String())
			en.okuriBuf.Reset()
			en.hasCandList = false
			en.candList = []string{}
			en.candIndex = 0
			en.cand = ""
			en.convMode = convGokan
			return output.String(), true
		case convGokan, convAbbrev:
			if en.cand == "" {
				en.resetConv()
				return output.String(), true
			} else {
				en.hasCandList = false
				en.candList = []string{}
				en.candIndex = 0
				en.cand = ""
				return output.String(), true
			}
		}
	}

	if en.candIndex >= candOffset && r != ' ' && r != 'x' {
		index := en.keyToCandIndex(r)
		if index < 0 {
			en.message = fmt.Sprintf(
				//"\"%c\" is not valid here!", r,
				"\"%c\" は有効なキーではありません！", r,
			)
			return output.String(), true
		}
		en.cand = en.candByIndex(index)

		flush()
		en.resetConv()
		return output.String(), true
	}

	if r == '\f' { // Ctrl-L
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
		if en.convMode == convAbbrev {
			output.WriteString(romaji.HanToZen(en.convBuf.String()))
			en.resetConv()
			return output.String(), true
		}

		if en.convMode != convNone {
			flush()
		}

		if en.inputMode != inputHira && en.inputMode != inputKata {
			en.inputMode = inputHira
		}

		en.resetConv()

		return output.String(), true
	}

	if r == termi.RuneEnter && en.convMode == convNone {
		if en.lineMode {
			if en.lineBuf.Len() > 0 {
				output.WriteString(en.lineBuf.String())
				en.lineBuf.Reset()
			} else {
				output.WriteRune(r)
			}
			return output.String(), true
		} else {
			output.WriteRune(r)
			return output.String(), false
		}
	}

	if en.inputMode == inputZen {
		zen, ok := romaji.ToZen[string(r)]
		if ok {
			if en.lineMode {
				en.lineBuf.WriteString(zen)
				return output.String(), true
			} else {
				output.WriteString(zen)
				return output.String(), false
			}
		} else {
			if en.lineMode {
				en.lineBuf.WriteRune(r)
				return output.String(), true
			} else {
				output.WriteRune(r)
				return output.String(), false
			}
		}
	}

	if en.inputMode == inputASCII {
		if en.lineMode {
			en.lineBuf.WriteRune(r)
			return output.String(), true
		} else {
			output.WriteRune(r)
			return output.String(), false
		}
	}

	// now in Hiragana or Katakana mode

	if r == 'L' {
		en.inputMode = inputZen
		en.inputBuf.Reset()

		if en.convMode != convNone {
			flush()
			en.resetConv()
		}

		return output.String(), true
	}

	if r == termi.RuneEnter && en.convMode != convNone {
		flush()
		en.resetConv()
		return output.String(), true
	}

	if r == termi.RuneEscape {
		en.inputMode = inputASCII
		en.inputBuf.Reset()

		if en.convMode != convNone {
			flush()
			en.resetConv()
		}

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

	if r == ' ' && en.convMode != convNone {
		if !en.hasCandList {
			if en.inputBuf.String() == "n" {
				if en.inputMode == inputHira {
					en.convBuf.WriteString("ん")
				} else if en.inputMode == inputKata {
					en.convBuf.WriteString("ン")
				}
				en.inputBuf.Reset()
			}

			gokan := en.convBuf.String()
			gokan = romaji.KataToHira(gokan)
			var err error
			if en.convMode == convOkuri {
				okuri := en.okuriBuf.String()
				okuri = romaji.KataToHira(okuri)
				en.candList, err = en.j.LookupOkuri(gokan, okuri)
			} else {
				en.candList, err = en.j.Lookup(gokan)
			}
			en.candIndex = 0
			if err != nil {
				en.message = fmt.Sprintf("%v", err)
				en.hasCandList = false
			} else {
				if len(en.candList) < 1 {
					en.message = "候補なし"
					return output.String(), true
				}
				en.hasCandList = true
			}
		} else {
			if en.candIndex < candOffset {
				en.candIndex++
				if en.candIndex >= len(en.candList) {
					en.candIndex = max(len(en.candList)-1, 0)
				}
			} else {
				if en.candIndex+len(candKeys) < len(en.candList) {
					en.candIndex += len(candKeys)
				}
			}
		}
		if en.candIndex < len(en.candList) {
			en.cand = en.candByIndex(en.candIndex)
		} else {
			en.cand = ""
		}
		return output.String(), true
	}

	if r == 'x' && en.hasCandList {
		if en.candIndex > candOffset {
			en.candIndex -= len(candKeys)
		} else {
			en.candIndex--
		}
		if en.candIndex < 0 {
			en.hasCandList = false
			en.candList = []string{}
			en.candIndex = 0
			en.cand = ""
			en.okuriBuf.Reset()
		} else if en.candIndex < len(en.candList) {
			en.cand = en.candByIndex(en.candIndex)
		}
		return output.String(), true
	}

	if r == '/' {
		if en.hasCandList {
			flush()
			en.resetConv()
		}
		en.convMode = convAbbrev
		return output.String(), true
	}

	if en.convMode == convAbbrev {
		if en.hasCandList {
			flush()
			en.resetConv()
		} else {
			en.convBuf.WriteRune(r)
			return output.String(), true
		}
	}

	if en.cand != "" {
		if en.convMode == convOkuri {
			if en.okuriBuf.Len() > 0 &&
				!strings.HasSuffix(en.okuriBuf.String(), "っ") &&
				!strings.HasSuffix(en.okuriBuf.String(), "ッ") {
				if en.lineMode {
					en.lineBuf.WriteString(en.cand)
					en.lineBuf.WriteString(en.okuriBuf.String())
				} else {
					output.WriteString(en.cand)
					output.WriteString(en.okuriBuf.String())
				}
				en.resetConv()
				en.inputBuf.Reset()
			}
		} else {
			if en.lineMode {
				en.lineBuf.WriteString(en.cand)
				en.lineBuf.WriteString(en.okuriBuf.String())
			} else {
				output.WriteString(en.cand)
				output.WriteString(en.okuriBuf.String())
			}
			en.resetConv()
		}
	}

	if r >= 'A' && r <= 'Z' {
		if en.inputBuf.String() == "n" {
			if en.inputMode == inputHira {
				en.convBuf.WriteString("ん")
			} else if en.inputMode == inputKata {
				en.convBuf.WriteString("ン")
			}
			en.inputBuf.Reset()
		}
		if en.convMode == convNone {
			en.convMode = convGokan
		} else if en.convMode == convGokan {
			en.convMode = convOkuri
		}
		r += 'a' - 'A'
	}

	kigou, ok := romaji.ToKigou[string(r)]
	if ok {
		if en.convMode != convNone && kigou == "ー" {
			if !en.hasCandList {
				en.convBuf.WriteString(kigou)
				return output.String(), true
			}
		}

		update := en.inputBuf.Len() > 0
		en.inputBuf.Reset()
		if en.lineMode {
			en.lineBuf.WriteString(kigou)
			update = true
		} else {
			output.WriteString(kigou)
		}
		return output.String(), update
	}

	if r < 'a' || r > 'z' {
		update := false
		if en.convMode != convNone {
			flush()
			en.resetConv()
			update = true
		}

		if en.lineMode {
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

		if en.convMode != convNone {
			flush()
			en.resetConv()
		}

		return output.String(), true
	}

	if r == 'q' {
		if en.convMode == convNone {
			if en.inputMode == inputHira {
				en.inputMode = inputKata
			} else if en.inputMode == inputKata {
				en.inputMode = inputHira
			} else {
				panic("q: invalid inputMode == " + string(en.inputMode))
			}
			return output.String(), true
		} else if en.convMode == convGokan {
			if en.inputMode == inputHira {
				kata := romaji.HiraToKata(en.convBuf.String())
				en.hasCandList = false
				en.candList = []string{kata}
				en.candIndex = 0
				en.cand = kata
			} else if en.inputMode == inputKata {
				hira := romaji.KataToHira(en.convBuf.String())
				en.hasCandList = false
				en.candList = []string{hira}
				en.candIndex = 0
				en.cand = hira
			} else {
				panic("q (conv): invalid inputMode == " +
					string(en.inputMode))
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
		} else if en.inputMode == inputKata {
			kana = "ッ"
		} else {
			panic("sokuon: invalid inputMode == " + string(en.inputMode))
			kana = ""
		}
		en.inputBuf.RemoveHead()
		sokuon = true
	} else if _, ok := romaji.IsN[en.inputBuf.String()]; ok {
		if en.inputMode == inputHira {
			kana = "ん"
		} else if en.inputMode == inputKata {
			kana = "ン"
		} else {
			panic("n: invalid inputMode == " + string(en.inputMode))
			kana = ""
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
		} else if en.inputMode == inputKata {
			k, ok = romaji.ToKata[lookup]
		} else {
			panic("kana: invalid inputMode == " + string(en.inputMode))
			kana = ""
		}
		if ok {
			kana = k
			en.inputBuf.Reset()
		}
	}

	if en.convMode == convNone {
		if kana != "" {
			if en.lineMode {
				en.lineBuf.WriteString(kana)
			} else {
				output.WriteString(kana)
			}
		}
		return output.String(), true
	} else if en.convMode == convGokan {
		if kana != "" {
			en.convBuf.WriteString(kana)
			en.hasCandList = false
			en.candList = []string{}
			en.candIndex = 0
			en.cand = ""
		}
		return output.String(), true
	} else if en.convMode == convOkuri {
		vowel, ok := vowels[kana]
		if ok {
			en.convBuf.WriteString(vowel)
			en.okuriBuf.WriteString(kana)
		} else if kana != "" {
			en.okuriBuf.WriteString(kana)
		} else {
			en.convBuf.WriteRune(r)
		}

		if en.okuriBuf.Len() < 1 {
			return output.String(), true
		}

		if sokuon {
			return output.String(), true
		}

		gokan := en.convBuf.String()
		gokan = romaji.KataToHira(gokan)
		okuri := en.okuriBuf.String()
		okuri = romaji.KataToHira(okuri)
		var err error
		en.candList, err = en.j.LookupOkuri(gokan, okuri)
		en.candIndex = 0
		if err != nil {
			en.message = fmt.Sprintf("%v", err)
			en.cand = ""
		} else {
			if en.candIndex < len(en.candList) {
				en.cand = en.candByIndex(en.candIndex)
				en.hasCandList = true
			} else {
				en.cand = ""
			}
		}
		return output.String(), true
	} else {
		panic("Process: invalid convMode == " + string(en.convMode))
		return output.String(), false
	}
}

func (en *Engine) Status() (string, bool) {
	if en.message != "" {
		return "SKK: " + en.message, true
	}

	if en.candIndex >= candOffset {
		s := new(strings.Builder)

		for i := 0; i < len(candKeys); i++ {
			if en.candIndex+i >= len(en.candList) {
				break
			}

			s.WriteRune(candKeys[i] + 'A' - 'a')
			s.WriteRune(':')

			cand := en.candByIndex(en.candIndex + i)
			s.WriteString(cand)
			s.WriteString("  ")
		}

		k := max(len(en.candList)-en.candIndex-len(candKeys), 0)
		s.WriteString(fmt.Sprintf("[残り %d]", k))

		return s.String(), false
	}

	s := new(strings.Builder)
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

	if en.convMode != convNone {
		if len(en.cand) > 0 {
			s.WriteRune('▼')
			s.WriteString(en.cand)
		} else {
			s.WriteRune('▽')
			var gokan string
			if en.convMode == convOkuri {
				gokan = en.convBuf.Substring(0, en.convBuf.Len()-1)
			} else {
				gokan = en.convBuf.String()
			}
			s.WriteString(gokan)
		}
	}

	if en.convMode == convOkuri && en.inputBuf.Len() > 0 {
		s.WriteRune('*')
	}
	s.WriteString(en.okuriBuf.String())
	s.WriteString(en.inputBuf.String())

	return s.String(), false
}
