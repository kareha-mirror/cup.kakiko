package skk

import (
	"fmt"

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

var romajiMode = RomajiDirect
var kanaBuffer = new(RuneBuf)
var kanaBufferString string

func Process(key termi.Key) string {
	switch key.Kind {
	case termi.KeyRune:
		switch romajiMode {
		case RomajiAlphabet:
			switch key.Rune {
			case '\n': // Ctrl-J
				romajiMode = RomajiHiragana
				kanaBufferString = ""
				kanaBuffer.Reset()
				return ""
			default:
				alphabet, ok := romaji.ToAlphabet[string(key.Rune)]
				if ok {
					return alphabet
				}
				return string(key.Rune)
			}
		case RomajiHiragana, RomajiKatakana:
			switch key.Rune {
			case termi.RuneBackspace, termi.RuneDelete:
				if kanaBuffer.Backspace() {
					kanaBufferString = kanaBuffer.String()
					return ""
				}
				return string(key.Rune)
			case 'l':
				romajiMode = RomajiDirect
				kanaBufferString = ""
				kanaBuffer.Reset()
				return ""
			case 'q':
				if romajiMode == RomajiHiragana {
					romajiMode = RomajiKatakana
				} else { // RomajiKatakana
					romajiMode = RomajiHiragana
				}
				kanaBufferString = ""
				kanaBuffer.Reset()
				return ""
			case 'L':
				romajiMode = RomajiAlphabet
				kanaBufferString = ""
				kanaBuffer.Reset()
				return ""
			case termi.RuneEscape:
				romajiMode = RomajiDirect
				kanaBufferString = ""
				kanaBuffer.Reset()
				return string(key.Rune)
			default:
				kanaBuffer.WriteRune(key.Rune)
				kanaBufferString = kanaBuffer.String()
				var kana string

				if romaji.IsSokuon[kanaBufferString] {
					if romajiMode == RomajiHiragana {
						kana = "っ"
					} else { // RomajiKatakana
						kana = "ッ"
					}
					kanaBuffer.RemoveHead()
					kanaBufferString = kanaBuffer.String()
				} else if romaji.IsN[kanaBufferString] {
					if romajiMode == RomajiHiragana {
						kana = "ん"
					} else { // RomajiKatakana
						kana = "ン"
					}
					kanaBuffer.RemoveHead()
					kanaBufferString = kanaBuffer.String()
				} else {
					lookup, ok := romaji.Aliases[kanaBufferString]
					if !ok {
						lookup = kanaBufferString
					}
					if romajiMode == RomajiHiragana {
						kana, ok = romaji.ToHiragana[lookup]
					} else { // RomajiKatakana
						kana, ok = romaji.ToKatakana[lookup]
					}
					if ok {
						kanaBufferString = kanaBuffer.String()
						kanaBuffer.Reset()
					}
				}

				return kana
			}
		default: // RomajiDirect
			switch key.Rune {
			case '\n': // Ctrl-J
				romajiMode = RomajiHiragana
				kanaBufferString = ""
				kanaBuffer.Reset()
				return ""
			default:
				return string(key.Rune)
			}
		}
	default:
		return key.Raw
	}
}

func Status() string {
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
	return fmt.Sprintf("[%s]%s", mark, kanaBufferString)
}
