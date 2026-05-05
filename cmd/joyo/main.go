package main

import (
	"os"
	"sort"
	"strings"

	"tea.kareha.org/cup/kakiko/internal/skkdic"
)

func load(
	path string,
	kanji map[string]string,
	okuri map[string]string,
) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, ";") {
			continue
		}

		space := strings.Index(line, " ")
		if space < 0 {
			continue
		}
		yomi := line[:space]
		cands := line[space+1:]

		if skkdic.HasOkuri(yomi) {
			prev, ok := okuri[yomi]
			if ok {
				okuri[yomi] = cands + prev[1:]
			} else {
				okuri[yomi] = cands
			}
		} else {
			prev, ok := kanji[yomi]
			if ok {
				kanji[yomi] = cands + prev[1:]
			} else {
				kanji[yomi] = cands
			}
		}
	}

	return nil
}

func sortKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func sortKeysReverse(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	return keys
}
func sortValue(v string) string {
	values := strings.Split(v[1:len(v)-1], "/")
	m := map[string]bool{}
	for _, value := range values {
		m[value] = true
	}
	keys := []string{}
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return "/" + strings.Join(keys, "/") + "/"
}

func main() {
	m := map[string]string{}

	for _, path := range os.Args[1:] {
		r, err := os.Open(path)
		if err != nil {
			panic(err)
		}

		err = skkdic.Load(r, m)
		if err != nil {
			panic(err)
		}
	}

	skkdic.Save(os.Stdout, m)
}
