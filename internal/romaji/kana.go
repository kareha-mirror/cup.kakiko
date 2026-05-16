package romaji

func HiraToKata(s string) string {
	var out []rune

	for _, r := range s {
		if r >= 'ぁ' && r <= 'ゖ' {
			r += 'ァ' - 'ぁ'
		}
		out = append(out, r)
	}

	return string(out)
}

func KataToHira(s string) string {
	var out []rune

	for _, r := range s {
		if r >= 'ァ' && r <= 'ヶ' {
			r -= 'ァ' - 'ぁ'
		}
		out = append(out, r)
	}

	return string(out)
}
