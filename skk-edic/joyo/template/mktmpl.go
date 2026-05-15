package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

func isKatakana(s string) bool {
	for _, r := range s {
		if !unicode.In(r, unicode.Katakana) {
			return false
		}
	}
	return true
}

func toHiragana(s string) string {
	var out []rune

	for _, r := range s {
		if r >= 'ァ' && r <= 'ヶ' {
			r -= 'ァ' - 'ぁ'
		}
		out = append(out, r)
	}

	return string(out)
}

func splitKanjis(s string) []string {
	kanji1 := ""
	kanji2 := ""
	for _, r := range s {
		if kanji1 == "" {
			kanji1 = string([]rune{r})
			continue
		}
		if r == '(' || r == ')' || r == '[' || r == ']' {
			continue
		}
		kanji2 = string([]rune{r})
		break
	}
	if kanji2 != "" {
		return []string{kanji1, kanji2}
	} else {
		return []string{kanji1}
	}
}

func main() {
	in, _ := os.Open(os.Args[1])
	defer in.Close()
	scanner := bufio.NewScanner(in)

	out, _ := os.Create(os.Args[2])
	defer out.Close()
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	lines := []string{}

	flush := func() {
		cols := strings.Split(lines[0], ",")
		kanjis := cols[0]
		writer.WriteString(fmt.Sprintf("; %%%%%% %s %%%%%%\n", kanjis))
		if len(cols) >= 2 {
			writer.WriteString(fmt.Sprintf("; %s:", cols[1]))
			if len(cols) >= 3 {
				writer.WriteString(fmt.Sprintf(" %s", cols[2]))
			}
			writer.WriteRune('\n')
		}
		if len(cols) >= 4 {
			writer.WriteString("; #")
			for i := 3; i < len(cols); i++ {
				writer.WriteRune(' ')
				writer.WriteString(cols[i])
			}
			writer.WriteRune('\n')
		}

		for i := 1; i < len(lines); i++ {
			cols = strings.Split(lines[i], ",")
			if len(cols) >= 2 {
				writer.WriteString(fmt.Sprintf("; %s:", cols[1]))
				if len(cols) >= 3 {
					writer.WriteString(fmt.Sprintf(" %s", cols[2]))
				}
				writer.WriteRune('\n')
			}
			if len(cols) >= 4 {
				writer.WriteString("; #")
				for i := 3; i < len(cols); i++ {
					writer.WriteRune(' ')
					writer.WriteString(cols[i])
				}
				writer.WriteRune('\n')
			}
		}

		writer.WriteRune('\n')

		for _, line := range lines {
			cols = strings.Split(line, ",")
			if len(cols) >= 2 {
				reading := cols[1]
				if isKatakana(reading) {
					list := splitKanjis(kanjis)
					if len(list) <= 1 {
						writer.WriteString(fmt.Sprintf("%s /%s/\n", toHiragana(reading), list[0]))
					} else {
						writer.WriteString(fmt.Sprintf("%s /%s/%s/\n", toHiragana(reading), list[0], list[1]))
					}
				}
			}
			if len(cols) >= 3 {
				words := strings.Split(cols[2], " ")
				for _, word := range words {
					writer.WriteString(fmt.Sprintf("? /%s/\n", word))
				}
			}
		}

		writer.WriteRune('\n')

		lines = lines[:0]
	}

	line := scanner.Text()
	lines = append(lines, line)
	for scanner.Scan() {
		line = scanner.Text()
		if !strings.HasPrefix(line, ",") {
			flush()
		}
		lines = append(lines, line)
	}
	flush()
}
