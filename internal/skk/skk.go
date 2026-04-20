package skk

import (
	"fmt"
	"strings"

	"tea.kareha.org/cup/termi"

	"tea.kareha.org/cup/kakiko/internal/romaji"
)

type RomajiMode int

const (
	RomajiDirect RomajiMode = iota
	RomajiHiragana
	RomajiKatakana
	RomajiAlphabet
)

type ConvMode int

const (
	ConvNone ConvMode = iota
	ConvStart
	ConvOkuri
	ConvEnglish
)

var romajiMode = RomajiDirect
var kanaBuilder = new(termi.StringBuilder)
var convMode = ConvNone
var convBuilder = new(termi.StringBuilder)
var hasConvList = false
var convList = []string{}
var convIndex = 0
var convCand = ""
var convOkuri = new(termi.StringBuilder)

var message = ""

var vowels = map[string]string{
	"あ": "a",
	"い": "i",
	"う": "u",
	"え": "e",
	"お": "o",
}

func resetConv() {
	convMode = ConvNone
	convBuilder.Reset()
	hasConvList = false
	convList = []string{}
	convIndex = 0
	convCand = ""
	convOkuri.Reset()
}

func Process(key termi.Key) string {
	switch key.Kind {
	case termi.KeyRune:
		r := key.Rune

		if r == termi.RuneBackspace || r == termi.RuneDelete {
			if kanaBuilder.RemoveTail() {
				return ""
			} else if convOkuri.RemoveTail() {
				return ""
			} else if convMode != ConvNone && convBuilder.RemoveTail() {
				if convBuilder.Len() < 1 {
					convMode = ConvNone
				}
				hasConvList = false
				convList = []string{}
				convIndex = 0
				convCand = ""
				return ""
			} else {
				return string(r)
			}
		}

		output := new(strings.Builder)

		flush := func() {
			var out string
			if convCand != "" {
				out = convCand + convOkuri.String()
			} else {
				out = convBuilder.String() + convOkuri.String()
			}
			output.WriteString(out)
		}

		// Ctrl-J
		if r == '\n' {
			if convMode != ConvNone {
				flush()
			}

			romajiMode = RomajiHiragana
			kanaBuilder.Reset()

			resetConv()

			return output.String()
		}

		if romajiMode == RomajiAlphabet {
			alphabet, ok := romaji.HankakuToZenkaku[string(r)]
			if ok {
				output.WriteString(alphabet)
			}
			return output.String()
		}

		if romajiMode == RomajiDirect {
			output.WriteRune(r)
			return output.String()
		}

		// now in Hiragana or Katakana mode

		if r == 'L' {
			romajiMode = RomajiAlphabet
			kanaBuilder.Reset()

			if convMode != ConvNone {
				flush()
				resetConv()
			}

			return output.String()
		}

		if r == termi.RuneEnter && convMode != ConvNone {
			flush()
			resetConv()
			return output.String()
		}

		if r == termi.RuneEnter {
			output.WriteRune(r)
			return output.String()
		}

		if r == termi.RuneEscape {
			romajiMode = RomajiDirect
			kanaBuilder.Reset()

			if convMode != ConvNone {
				flush()
				resetConv()
			}

			output.WriteRune(r)

			return output.String()
		}

		if r < ' ' {
			output.WriteRune(r)
			return output.String()
		}

		if r == ' ' && convMode != ConvNone {
			if !hasConvList {
				var err error
				if convMode == ConvOkuri {
					convList, err = LookupOkuri(
						convBuilder.String(), convOkuri.String(),
					)
				} else {
					convList, err = Lookup(convBuilder.String())
				}
				convIndex = 0
				if err != nil {
					message = fmt.Sprintf("%v", err)
					hasConvList = false
				} else {
					if len(convList) < 1 {
						message = "SKK: 候補なし"
						return output.String()
					}
					hasConvList = true
				}
			} else {
				convIndex++
				if convIndex >= len(convList) {
					convIndex = max(len(convList)-1, 0)
				}
			}
			if convIndex < len(convList) {
				convCand = convList[convIndex]
				semicolon := strings.Index(convCand, ";")
				if semicolon >= 0 {
					convCand = convCand[:semicolon]
				}
			} else {
				convCand = ""
			}
			return output.String()
		}

		if r == 'x' && hasConvList {
			convIndex--
			if convIndex < 0 {
				hasConvList = false
				convList = []string{}
				convIndex = 0
				convCand = ""
				convOkuri.Reset()
			} else {
				if convIndex < len(convList) {
					convCand = convList[convIndex]
					semicolon := strings.Index(convCand, ";")
					if semicolon >= 0 {
						convCand = convCand[:semicolon]
					}
				}
			}
			return output.String()
		}

		if r == '/' {
			if hasConvList {
				flush()
				resetConv()
			}
			convMode = ConvEnglish
			return output.String()
		}

		if convMode == ConvEnglish {
			if hasConvList {
				flush()
				resetConv()
			} else {
				convBuilder.WriteRune(r)
				return output.String()
			}
		}

		if convCand != "" {
			if convMode == ConvOkuri {
				if convOkuri.Len() > 0 &&
					!strings.HasSuffix(convOkuri.String(), "っ") &&
					!strings.HasSuffix(convOkuri.String(), "ッ") {
					output.WriteString(convCand)
					output.WriteString(convOkuri.String())
					resetConv()
					kanaBuilder.Reset()
				}
			} else {
				output.WriteString(convCand)
				output.WriteString(convOkuri.String())
				resetConv()
			}
		}

		if r >= 'A' && r <= 'Z' {
			if convMode == ConvNone {
				convMode = ConvStart
			} else if convMode == ConvStart {
				convMode = ConvOkuri
			}
			r += 'a' - 'A'
		}

		kigou, ok := romaji.ToKigou[string(r)]
		if ok {
			if convMode != ConvNone && kigou == "ー" {
				if !hasConvList {
					convBuilder.WriteString(kigou)
					return output.String()
				}
			}

			kanaBuilder.Reset()
			output.WriteString(kigou)
			return output.String()
		}

		if r < 'a' && r > 'z' {
			if convMode != ConvNone {
				flush()
				resetConv()
			}

			output.WriteRune(r)
			return output.String()
		}

		if r == 'l' {
			romajiMode = RomajiDirect
			kanaBuilder.Reset()

			if convMode != ConvNone {
				flush()
				resetConv()
			}

			return output.String()
		}

		if r == 'q' {
			if convMode == ConvNone {
				if romajiMode == RomajiHiragana {
					romajiMode = RomajiKatakana
				} else if romajiMode == RomajiKatakana {
					romajiMode = RomajiHiragana
				} else {
					panic("q: invalid romajiMode == " + string(romajiMode))
				}
				return output.String()
			} else if convMode == ConvStart {
				if romajiMode == RomajiHiragana {
					kata := romaji.HiraToKata(convBuilder.String())
					hasConvList = false
					convList = []string{kata}
					convIndex = 0
					convCand = kata
				} else if romajiMode == RomajiKatakana {
					hira := romaji.KataToHira(convBuilder.String())
					hasConvList = false
					convList = []string{hira}
					convIndex = 0
					convCand = hira
				} else {
					panic("q (conv): invalid romajiMode == " +
						string(romajiMode))
				}
				return output.String()
			}
		}

		kanaBuilder.WriteRune(r)

		var kana string
		if _, ok := romaji.IsSokuon[kanaBuilder.String()]; ok {
			if romajiMode == RomajiHiragana {
				kana = "っ"
			} else if romajiMode == RomajiKatakana {
				kana = "ッ"
			} else {
				panic("sokuon: invalid romajiMode == " + string(romajiMode))
				kana = ""
			}
			kanaBuilder.RemoveHead()
		} else if _, ok := romaji.IsN[kanaBuilder.String()]; ok {
			if romajiMode == RomajiHiragana {
				kana = "ん"
			} else if romajiMode == RomajiKatakana {
				kana = "ン"
			} else {
				panic("n: invalid romajiMode == " + string(romajiMode))
				kana = ""
			}
			kanaBuilder.RemoveHead()
		} else {
			lookup := kanaBuilder.String()
			alias, ok := romaji.Aliases[lookup]
			if ok {
				lookup = alias
			}

			var k string
			if romajiMode == RomajiHiragana {
				k, ok = romaji.ToHiragana[lookup]
			} else if romajiMode == RomajiKatakana {
				k, ok = romaji.ToKatakana[lookup]
			} else {
				panic("kana: invalid romajiMode == " + string(romajiMode))
				kana = ""
			}
			if ok {
				kana = k
				kanaBuilder.Reset()
			}
		}

		if convMode == ConvNone {
			if kana != "" {
				output.WriteString(kana)
			}
			return output.String()
		} else if convMode == ConvStart {
			if kana != "" {
				convBuilder.WriteString(kana)
				hasConvList = false
				convList = []string{}
				convIndex = 0
				convCand = ""
			}
			return output.String()
		} else if convMode == ConvOkuri {
			vowel, ok := vowels[kana]
			if ok {
				convBuilder.WriteString(vowel)
				convOkuri.WriteString(kana)
			} else if kana != "" {
				convOkuri.WriteString(kana)
			} else {
				convBuilder.WriteRune(r)
			}

			var err error
			convList, err = LookupOkuri(
				convBuilder.String(), convOkuri.String(),
			)
			convIndex = 0
			if err != nil {
				message = fmt.Sprintf("%v", err)
				convCand = ""
			} else {
				if convIndex < len(convList) {
					convCand = convList[convIndex]
					semicolon := strings.Index(convCand, ";")
					if semicolon >= 0 {
						convCand = convCand[:semicolon]
					}
					hasConvList = true
				} else {
					convCand = ""
				}
			}

			return output.String()
		} else {
			panic("Process: invalid convMode == " + string(convMode))
			return output.String()
		}
	default:
		return key.Raw
	}
}

func Status() string {
	if message != "" {
		m := message
		message = ""
		return m
	}

	var mark string
	switch romajiMode {
	case RomajiHiragana:
		mark = "あ"
	case RomajiKatakana:
		mark = "ア"
	case RomajiAlphabet:
		mark = "ａＡ"
	default: // RomajiDirect
		mark = "aA"
	}

	head := ""
	if convMode != ConvNone {
		if len(convCand) > 0 {
			head = "▼" + convCand
		} else {
			head = "▽" + convBuilder.String()
		}
	}

	var buf string
	if len(convOkuri.String()) > 0 {
		buf = convOkuri.String()
	} else {
		buf = kanaBuilder.String()
	}

	return fmt.Sprintf("[%s]%s%s", mark, head, buf)
}
