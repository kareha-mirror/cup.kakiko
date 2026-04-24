package skk

import (
	"fmt"
	"strings"

	"tea.kareha.org/cup/termi"

	"tea.kareha.org/cup/kakiko/internal/romaji"
)

type romajiMode int

const (
	romajiDirect romajiMode = iota
	romajiHiragana
	romajiKatakana
	romajiAlphabet
)

type convMode int

const (
	convNone convMode = iota
	convStart
	convOkuri
	convEnglish
)

type Engine struct {
	j Jisyo

	romajiMode  romajiMode
	kanaBuilder *termi.StringBuilder
	convMode    convMode
	convBuilder *termi.StringBuilder
	hasConvList bool
	convList    []string
	convIndex   int
	convCand    string
	convOkuri   *termi.StringBuilder

	lineMode   bool
	lineBuffer *termi.StringBuilder
	linePass   bool

	message string
}

func NewEngine(path string) *Engine {
	en := &Engine{
		j: NewCDBJisyo(path),

		romajiMode:  romajiDirect,
		kanaBuilder: new(termi.StringBuilder),
		convMode:    convNone,
		convBuilder: new(termi.StringBuilder),
		hasConvList: false,
		convList:    []string{},
		convIndex:   0,
		convCand:    "",
		convOkuri:   new(termi.StringBuilder),

		lineMode:   false,
		lineBuffer: new(termi.StringBuilder),
		linePass:   false,

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
			if en.kanaBuilder.RemoveTail() {
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
			} else if en.lineMode && en.lineBuffer.RemoveTail() {
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
				en.lineBuffer.WriteString(out)
			} else {
				output.WriteString(out)
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
				output.WriteString(en.lineBuffer.String())
				en.lineBuffer.Reset()
			} else {
				en.lineMode = true
			}

			return output.String(), true
		}
		en.linePass = false

		// Ctrl-J
		if r == '\n' {
			if en.convMode != convNone {
				flush()
			}

			en.romajiMode = romajiHiragana
			en.kanaBuilder.Reset()

			en.resetConv()

			return output.String(), true
		}

		if en.romajiMode == romajiAlphabet {
			alphabet, ok := romaji.HankakuToZenkaku[string(r)]
			if ok {
				if en.lineMode {
					en.lineBuffer.WriteString(alphabet)
					return output.String(), true
				} else {
					output.WriteString(alphabet)
				}
			}
			return output.String(), false
		}

		if r == termi.RuneEnter && en.convMode == convNone {
			if en.lineMode {
				if en.lineBuffer.Len() > 0 {
					output.WriteString(en.lineBuffer.String())
					en.lineBuffer.Reset()
				} else {
					output.WriteRune(r)
				}
				return output.String(), true
			} else {
				output.WriteRune(r)
				return output.String(), false
			}
		}

		if en.romajiMode == romajiDirect {
			if en.lineMode {
				en.lineBuffer.WriteRune(r)
				return output.String(), true
			} else {
				output.WriteRune(r)
				return output.String(), false
			}
		}

		// now in Hiragana or Katakana mode

		if r == 'L' {
			en.romajiMode = romajiAlphabet
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
			en.romajiMode = romajiDirect
			en.kanaBuilder.Reset()

			if en.convMode != convNone {
				flush()
				en.resetConv()
			}

			if en.lineMode {
				output.WriteString(en.lineBuffer.String())
				en.lineBuffer.Reset()
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
				var err error
				if en.convMode == convOkuri {
					en.convList, err = en.j.LookupOkuri(
						en.convBuilder.String(), en.convOkuri.String(),
					)
				} else {
					en.convList, err = en.j.Lookup(en.convBuilder.String())
				}
				en.convIndex = 0
				if err != nil {
					en.message = fmt.Sprintf("%v", err)
					en.hasConvList = false
				} else {
					if len(en.convList) < 1 {
						en.message = "SKK: 候補なし"
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
			en.convMode = convEnglish
			return output.String(), true
		}

		if en.convMode == convEnglish {
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
						en.lineBuffer.WriteString(en.convCand)
						en.lineBuffer.WriteString(en.convOkuri.String())
					} else {
						output.WriteString(en.convCand)
						output.WriteString(en.convOkuri.String())
					}
					en.resetConv()
					en.kanaBuilder.Reset()
				}
			} else {
				if en.lineMode {
					en.lineBuffer.WriteString(en.convCand)
					en.lineBuffer.WriteString(en.convOkuri.String())
				} else {
					output.WriteString(en.convCand)
					output.WriteString(en.convOkuri.String())
				}
				en.resetConv()
			}
		}

		if r >= 'A' && r <= 'Z' {
			if en.convMode == convNone {
				en.convMode = convStart
			} else if en.convMode == convStart {
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
				en.lineBuffer.WriteString(kigou)
				update = true
			} else {
				output.WriteString(kigou)
			}
			return output.String(), update
		}

		if r < 'a' && r > 'z' {
			update := false
			if en.convMode != convNone {
				flush()
				en.resetConv()
				update = true
			}

			if en.lineMode {
				en.lineBuffer.WriteRune(r)
				update = true
			} else {
				output.WriteRune(r)
			}
			return output.String(), update
		}

		if r == 'l' {
			en.romajiMode = romajiDirect
			en.kanaBuilder.Reset()

			if en.convMode != convNone {
				flush()
				en.resetConv()
			}

			return output.String(), true
		}

		if r == 'q' {
			if en.convMode == convNone {
				if en.romajiMode == romajiHiragana {
					en.romajiMode = romajiKatakana
				} else if en.romajiMode == romajiKatakana {
					en.romajiMode = romajiHiragana
				} else {
					panic("q: invalid romajiMode == " + string(en.romajiMode))
				}
				return output.String(), true
			} else if en.convMode == convStart {
				if en.romajiMode == romajiHiragana {
					kata := romaji.HiraToKata(en.convBuilder.String())
					en.hasConvList = false
					en.convList = []string{kata}
					en.convIndex = 0
					en.convCand = kata
				} else if en.romajiMode == romajiKatakana {
					hira := romaji.KataToHira(en.convBuilder.String())
					en.hasConvList = false
					en.convList = []string{hira}
					en.convIndex = 0
					en.convCand = hira
				} else {
					panic("q (conv): invalid romajiMode == " +
						string(en.romajiMode))
				}
				return output.String(), true
			}
		}

		en.kanaBuilder.WriteRune(r)

		var kana string
		if _, ok := romaji.IsSokuon[en.kanaBuilder.String()]; ok {
			if en.romajiMode == romajiHiragana {
				kana = "っ"
			} else if en.romajiMode == romajiKatakana {
				kana = "ッ"
			} else {
				panic("sokuon: invalid romajiMode == " + string(en.romajiMode))
				kana = ""
			}
			en.kanaBuilder.RemoveHead()
		} else if _, ok := romaji.IsN[en.kanaBuilder.String()]; ok {
			if en.romajiMode == romajiHiragana {
				kana = "ん"
			} else if en.romajiMode == romajiKatakana {
				kana = "ン"
			} else {
				panic("n: invalid romajiMode == " + string(en.romajiMode))
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
			if en.romajiMode == romajiHiragana {
				k, ok = romaji.ToHiragana[lookup]
			} else if en.romajiMode == romajiKatakana {
				k, ok = romaji.ToKatakana[lookup]
			} else {
				panic("kana: invalid romajiMode == " + string(en.romajiMode))
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
					en.lineBuffer.WriteString(kana)
				} else {
					output.WriteString(kana)
				}
			}
			return output.String(), true
		} else if en.convMode == convStart {
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

			var err error
			en.convList, err = en.j.LookupOkuri(
				en.convBuilder.String(), en.convOkuri.String(),
			)
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
		return m
	}

	var mark string
	switch en.romajiMode {
	case romajiHiragana:
		mark = "あ"
	case romajiKatakana:
		mark = "ア"
	case romajiAlphabet:
		mark = "ａＡ"
	default: // romajiDirect
		mark = "aA"
	}

	head := ""
	if en.convMode != convNone {
		if len(en.convCand) > 0 {
			head = "▼" + en.convCand
		} else {
			head = "▽" + en.convBuilder.String()
		}
	}

	var buf string
	if len(en.convOkuri.String()) > 0 {
		buf = en.convOkuri.String()
	} else {
		buf = en.kanaBuilder.String()
	}

	if en.lineMode {
		return fmt.Sprintf(
			"[%s]%s:%s%s", mark, en.lineBuffer.String(), head, buf,
		)
	} else {
		return fmt.Sprintf("[%s]%s%s", mark, head, buf)
	}
}
