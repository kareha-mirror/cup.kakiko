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
	convBody
	convOkuri
	convAbbrev
)

type Engine struct {
	j Jisyo

	inputMode   inputMode
	kanaBuilder *termi.StringBuilder
	convMode    convMode
	convBuilder *termi.StringBuilder
	hasConvList bool
	convList    []string
	convIndex   int
	convCand    string
	convOkuri   *termi.StringBuilder

	lineMode    bool
	lineBuilder *termi.StringBuilder
	linePass    bool

	message string
}

func NewEngine(path string) *Engine {
	en := &Engine{
		j: NewCDBJisyo(path),

		inputMode:   inputASCII,
		kanaBuilder: new(termi.StringBuilder),
		convMode:    convNone,
		convBuilder: new(termi.StringBuilder),
		hasConvList: false,
		convList:    []string{},
		convIndex:   0,
		convCand:    "",
		convOkuri:   new(termi.StringBuilder),

		lineMode:    false,
		lineBuilder: new(termi.StringBuilder),
		linePass:    false,

		message: "",
	}

	return en
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
	en.convBuilder.Reset()
	en.hasConvList = false
	en.convList = []string{}
	en.convIndex = 0
	en.convCand = ""
	en.convOkuri.Reset()
}

func (en *Engine) Process(key termi.Key) (string, bool) {
	switch key.Kind {
	case termi.KeyRune:
		r := key.Rune

		if r == termi.RuneBackspace || r == termi.RuneDelete {
			if en.kanaBuilder.Len() > 0 {
				en.kanaBuilder.Reset()
				return "", true
			} else if en.convOkuri.RemoveTail() {
				return "", true
			} else if en.convMode != convNone && en.convBuilder.RemoveTail() {
				if en.convBuilder.Len() < 1 {
					en.convMode = convNone
				}
				en.hasConvList = false
				en.convList = []string{}
				en.convIndex = 0
				en.convCand = ""
				return "", true
			} else if en.lineMode && en.lineBuilder.RemoveTail() {
				return "", true
			} else {
				return string(r), false
			}
		}

		output := new(strings.Builder)

		flush := func() {
			var out string
			if en.convCand != "" {
				out = en.convCand + en.convOkuri.String()
			} else {
				out = en.convBuilder.String() + en.convOkuri.String()
			}
			if en.lineMode {
				en.lineBuilder.WriteString(out)
			} else {
				output.WriteString(out)
			}
		}

		// Ctrl-G
		if r == '\a' {
			if en.convMode == convNone {
				if en.kanaBuilder.Len() > 0 {
					en.kanaBuilder.Reset()
					return output.String(), true
				}
			} else if en.convMode == convOkuri {
				en.kanaBuilder.Reset()
				en.convBuilder.RemoveTail()
				en.convBuilder.WriteString(en.convOkuri.String())
				en.convOkuri.Reset()
				en.convCand = ""
				en.convIndex = 0
				en.convMode = convBody
				return output.String(), true
			} else if en.convMode == convBody || en.convMode == convAbbrev {
				if en.convCand == "" {
					en.resetConv()
					return output.String(), true
				} else {
					en.convCand = ""
					en.convIndex = 0
					return output.String(), true
				}
			}
		}

		// Ctrl-L
		if r == '\f' {
			if en.linePass {
				en.linePass = false
				output.WriteRune(r)
			} else {
				en.linePass = true
			}

			if en.lineMode {
				en.lineMode = false
				flush()
				output.WriteString(en.lineBuilder.String())
				en.lineBuilder.Reset()
			} else {
				en.lineMode = true
			}

			return output.String(), true
		}
		en.linePass = false

		// Ctrl-J
		if r == '\n' {
			if en.convMode == convAbbrev {
				output.WriteString(romaji.HanToZen(en.convBuilder.String()))
				en.resetConv()
				return output.String(), true
			}

			if en.convMode != convNone {
				flush()
			}

			if en.inputMode != inputHira && en.inputMode != inputKata {
				en.inputMode = inputHira
			}
			en.kanaBuilder.Reset()

			en.resetConv()

			return output.String(), true
		}

		if en.inputMode == inputZen {
			alphabet, ok := romaji.ToZen[string(r)]
			if ok {
				if en.lineMode {
					en.lineBuilder.WriteString(alphabet)
					return output.String(), true
				} else {
					output.WriteString(alphabet)
				}
			}
			return output.String(), false
		}

		if r == termi.RuneEnter && en.convMode == convNone {
			if en.lineMode {
				if en.lineBuilder.Len() > 0 {
					output.WriteString(en.lineBuilder.String())
					en.lineBuilder.Reset()
				} else {
					output.WriteRune(r)
				}
				return output.String(), true
			} else {
				output.WriteRune(r)
				return output.String(), false
			}
		}

		if en.inputMode == inputASCII {
			if en.lineMode {
				en.lineBuilder.WriteRune(r)
				return output.String(), true
			} else {
				output.WriteRune(r)
				return output.String(), false
			}
		}

		// now in Hiragana or Katakana mode

		if r == 'L' {
			en.inputMode = inputZen
			en.kanaBuilder.Reset()

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
			en.kanaBuilder.Reset()

			if en.convMode != convNone {
				flush()
				en.resetConv()
			}

			if en.lineMode {
				output.WriteString(en.lineBuilder.String())
				en.lineBuilder.Reset()
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
			if !en.hasConvList {
				if en.kanaBuilder.String() == "n" {
					if en.inputMode == inputHira {
						en.convBuilder.WriteString("ん")
					} else if en.inputMode == inputKata {
						en.convBuilder.WriteString("ン")
					}
					en.kanaBuilder.Reset()
				}

				body := en.convBuilder.String()
				body = romaji.KataToHira(body)
				var err error
				if en.convMode == convOkuri {
					okuri := en.convOkuri.String()
					okuri = romaji.KataToHira(okuri)
					en.convList, err = en.j.LookupOkuri(body, okuri)
				} else {
					en.convList, err = en.j.Lookup(body)
				}
				en.convIndex = 0
				if err != nil {
					en.message = fmt.Sprintf("%v", err)
					en.hasConvList = false
				} else {
					if len(en.convList) < 1 {
						en.message = "候補なし"
						return output.String(), true
					}
					en.hasConvList = true
				}
			} else {
				en.convIndex++
				if en.convIndex >= len(en.convList) {
					en.convIndex = max(len(en.convList)-1, 0)
				}
			}
			if en.convIndex < len(en.convList) {
				en.convCand = en.convList[en.convIndex]
				semicolon := strings.Index(en.convCand, ";")
				if semicolon >= 0 {
					en.convCand = en.convCand[:semicolon]
				}
			} else {
				en.convCand = ""
			}
			return output.String(), true
		}

		if r == 'x' && en.hasConvList {
			en.convIndex--
			if en.convIndex < 0 {
				en.hasConvList = false
				en.convList = []string{}
				en.convIndex = 0
				en.convCand = ""
				en.convOkuri.Reset()
			} else {
				if en.convIndex < len(en.convList) {
					en.convCand = en.convList[en.convIndex]
					semicolon := strings.Index(en.convCand, ";")
					if semicolon >= 0 {
						en.convCand = en.convCand[:semicolon]
					}
				}
			}
			return output.String(), true
		}

		if r == '/' {
			if en.hasConvList {
				flush()
				en.resetConv()
			}
			en.convMode = convAbbrev
			return output.String(), true
		}

		if en.convMode == convAbbrev {
			if en.hasConvList {
				flush()
				en.resetConv()
			} else {
				en.convBuilder.WriteRune(r)
				return output.String(), true
			}
		}

		if en.convCand != "" {
			if en.convMode == convOkuri {
				if en.convOkuri.Len() > 0 &&
					!strings.HasSuffix(en.convOkuri.String(), "っ") &&
					!strings.HasSuffix(en.convOkuri.String(), "ッ") {
					if en.lineMode {
						en.lineBuilder.WriteString(en.convCand)
						en.lineBuilder.WriteString(en.convOkuri.String())
					} else {
						output.WriteString(en.convCand)
						output.WriteString(en.convOkuri.String())
					}
					en.resetConv()
					en.kanaBuilder.Reset()
				}
			} else {
				if en.lineMode {
					en.lineBuilder.WriteString(en.convCand)
					en.lineBuilder.WriteString(en.convOkuri.String())
				} else {
					output.WriteString(en.convCand)
					output.WriteString(en.convOkuri.String())
				}
				en.resetConv()
			}
		}

		if r >= 'A' && r <= 'Z' {
			if en.kanaBuilder.String() == "n" {
				if en.inputMode == inputHira {
					en.convBuilder.WriteString("ん")
				} else if en.inputMode == inputKata {
					en.convBuilder.WriteString("ン")
				}
				en.kanaBuilder.Reset()
			}
			if en.convMode == convNone {
				en.convMode = convBody
			} else if en.convMode == convBody {
				en.convMode = convOkuri
			}
			r += 'a' - 'A'
		}

		kigou, ok := romaji.ToKigou[string(r)]
		if ok {
			if en.convMode != convNone && kigou == "ー" {
				if !en.hasConvList {
					en.convBuilder.WriteString(kigou)
					return output.String(), true
				}
			}

			update := en.kanaBuilder.Len() > 0
			en.kanaBuilder.Reset()
			if en.lineMode {
				en.lineBuilder.WriteString(kigou)
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
				en.lineBuilder.WriteRune(r)
				update = true
			} else {
				output.WriteRune(r)
			}
			return output.String(), update
		}

		if r == 'l' {
			en.inputMode = inputASCII
			en.kanaBuilder.Reset()

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
			} else if en.convMode == convBody {
				if en.inputMode == inputHira {
					kata := romaji.HiraToKata(en.convBuilder.String())
					en.hasConvList = false
					en.convList = []string{kata}
					en.convIndex = 0
					en.convCand = kata
				} else if en.inputMode == inputKata {
					hira := romaji.KataToHira(en.convBuilder.String())
					en.hasConvList = false
					en.convList = []string{hira}
					en.convIndex = 0
					en.convCand = hira
				} else {
					panic("q (conv): invalid inputMode == " +
						string(en.inputMode))
				}
				return output.String(), true
			}
		}

		en.kanaBuilder.WriteRune(r)

		var kana string
		sokuon := false
		if _, ok := romaji.IsSokuon[en.kanaBuilder.String()]; ok {
			if en.inputMode == inputHira {
				kana = "っ"
			} else if en.inputMode == inputKata {
				kana = "ッ"
			} else {
				panic("sokuon: invalid inputMode == " + string(en.inputMode))
				kana = ""
			}
			en.kanaBuilder.RemoveHead()
			sokuon = true
		} else if _, ok := romaji.IsN[en.kanaBuilder.String()]; ok {
			if en.inputMode == inputHira {
				kana = "ん"
			} else if en.inputMode == inputKata {
				kana = "ン"
			} else {
				panic("n: invalid inputMode == " + string(en.inputMode))
				kana = ""
			}
			en.kanaBuilder.RemoveHead()
		} else {
			lookup := en.kanaBuilder.String()
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
				en.kanaBuilder.Reset()
			}
		}

		if en.convMode == convNone {
			if kana != "" {
				if en.lineMode {
					en.lineBuilder.WriteString(kana)
				} else {
					output.WriteString(kana)
				}
			}
			return output.String(), true
		} else if en.convMode == convBody {
			if kana != "" {
				en.convBuilder.WriteString(kana)
				en.hasConvList = false
				en.convList = []string{}
				en.convIndex = 0
				en.convCand = ""
			}
			return output.String(), true
		} else if en.convMode == convOkuri {
			vowel, ok := vowels[kana]
			if ok {
				en.convBuilder.WriteString(vowel)
				en.convOkuri.WriteString(kana)
			} else if kana != "" {
				en.convOkuri.WriteString(kana)
			} else {
				en.convBuilder.WriteRune(r)
			}

			if en.convOkuri.Len() < 1 {
				return output.String(), true
			}

			if sokuon {
				return output.String(), true
			}

			body := en.convBuilder.String()
			body = romaji.KataToHira(body)
			okuri := en.convOkuri.String()
			okuri = romaji.KataToHira(okuri)
			var err error
			en.convList, err = en.j.LookupOkuri(body, okuri)
			en.convIndex = 0
			if err != nil {
				en.message = fmt.Sprintf("%v", err)
				en.convCand = ""
			} else {
				if en.convIndex < len(en.convList) {
					en.convCand = en.convList[en.convIndex]
					semicolon := strings.Index(en.convCand, ";")
					if semicolon >= 0 {
						en.convCand = en.convCand[:semicolon]
					}
					en.hasConvList = true
				} else {
					en.convCand = ""
				}
			}
			return output.String(), true
		} else {
			panic("Process: invalid convMode == " + string(en.convMode))
			return output.String(), false
		}
	default:
		return key.Raw, false
	}
}

func (en *Engine) Status() string {
	if en.message != "" {
		m := en.message
		en.message = ""
		return "SKK: " + m
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
	default: // inputASCII
		panic("Status: invalid inputMode == " + string(en.inputMode))
	}
	s.WriteRune(')')

	if en.lineMode {
		s.WriteString(en.lineBuilder.String())
		s.WriteRune(':')
	}

	if en.convMode != convNone {
		if len(en.convCand) > 0 {
			s.WriteRune('▼')
			s.WriteString(en.convCand)
		} else {
			s.WriteRune('▽')
			if en.convMode == convOkuri {
				s.WriteString(
					en.convBuilder.Substring(0, en.convBuilder.Len()-1),
				)
			} else {
				s.WriteString(en.convBuilder.String())
			}
		}
	}

	star := false
	if len(en.convOkuri.String()) > 0 {
		if en.convCand == "" {
			s.WriteRune('*')
			star = true
		}
		s.WriteString(en.convOkuri.String())
	}
	if en.kanaBuilder.Len() > 0 {
		if en.convMode == convOkuri && !star {
			s.WriteRune('*')
		}
		s.WriteString(en.kanaBuilder.String())
	}

	if len(en.convCand) > 0 {
		k := len(en.convList) - en.convIndex - 1
		p := "+"
		if k < 1 {
			p = ""
		}
		s.WriteString(fmt.Sprintf("  [残り %d%s]", k, p))
	}

	return s.String()
}
