package romaji

import (
	"strings"
)

var hiraToKata = map[string]string{
	// あ
	"あ": "ア",
	"い": "イ",
	"う": "ウ",
	"え": "エ",
	"お": "オ",
	// か
	"か": "カ",
	"き": "キ",
	"く": "ク",
	"け": "ケ",
	"こ": "コ",
	// さ
	"さ": "サ",
	"し": "シ",
	"す": "ス",
	"せ": "セ",
	"そ": "ソ",
	// た
	"た": "タ",
	"ち": "チ",
	"つ": "ツ",
	"て": "テ",
	"と": "ト",
	// な
	"な": "ナ",
	"に": "ニ",
	"ぬ": "ヌ",
	"ね": "ネ",
	"の": "ノ",
	// は
	"は": "ハ",
	"ひ": "ヒ",
	"ふ": "フ",
	"へ": "ヘ",
	"ほ": "ホ",
	// ま
	"ま": "マ",
	"み": "ミ",
	"む": "ム",
	"め": "メ",
	"も": "モ",
	// や
	"や": "ヤ",
	"ゆ": "ユ",
	"よ": "ヨ",
	// ら
	"ら": "ラ",
	"り": "リ",
	"る": "ル",
	"れ": "レ",
	"ろ": "ロ",
	// わ
	"わ": "ワ",
	"ゐ": "ヰ",
	"ゑ": "ヱ",
	"を": "ヲ",
	// ん
	"ん": "ン",
	// が
	"が": "ガ",
	"ぎ": "ギ",
	"ぐ": "グ",
	"げ": "ゲ",
	"ご": "ゴ",
	// ざ
	"ざ": "ザ",
	"じ": "ジ",
	"ず": "ズ",
	"ぜ": "ゼ",
	"ぞ": "ゾ",
	// だ
	"だ": "ダ",
	"ぢ": "ヂ",
	"づ": "ヅ",
	"で": "デ",
	"ど": "ド",
	// ば
	"ば": "バ",
	"び": "ビ",
	"ぶ": "ブ",
	"べ": "ベ",
	"ぼ": "ボ",
	// ぱ
	"ぱ": "パ",
	"ぴ": "ピ",
	"ぷ": "プ",
	"ぺ": "ペ",
	"ぽ": "ポ",
	// ゔ
	"ゔ": "ヴ",
	// ぁ
	"ぁ": "ァ",
	"ぃ": "ィ",
	"ぅ": "ゥ",
	"ぇ": "ェ",
	"ぉ": "ォ",
	// っ
	"っ": "ッ",
	// ゃ
	"ゃ": "ャ",
	"ゅ": "ュ",
	"ょ": "ョ",
	// ゎ
	"ゎ": "ヮ",
}

var kataToHira = make(map[string]string, len(hiraToKata))

func init() {
	for k, v := range hiraToKata {
		kataToHira[v] = k
	}
}

func HiraToKata(s string) string {
	buf := strings.Builder{}

	for _, r := range s {
		k, ok := hiraToKata[string(r)]
		if ok {
			buf.WriteString(k)
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

func KataToHira(s string) string {
	buf := strings.Builder{}

	for _, r := range s {
		h, ok := kataToHira[string(r)]
		if ok {
			buf.WriteString(h)
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
