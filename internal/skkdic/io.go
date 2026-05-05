package skkdic

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

func HasOkuri(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i > 0 && i == len(s)-1 {
			if r >= 'a' || r <= 'z' {
				return true
			}
		} else {
			if r < 0x3041 || r > 0x3096 {
				return false
			}
		}
	}
	return false
}

func Load(r io.Reader, m map[string]string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if line[0] == ';' {
			continue
		}

		space := strings.Index(line, " ")
		if space < 0 {
			continue
		}
		if space+1 >= len(line) {
			continue
		}
		reading := line[:space]
		cands := line[space+1:]

		prev, ok := m[reading]
		if ok {
			m[reading] = cands + prev[1:]
		} else {
			m[reading] = cands
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

func Save(w io.Writer, m map[string]string) {
	kanji := map[string]string{}
	okuri := map[string]string{}

	for reading := range m {
		if HasOkuri(reading) {
			okuri[reading] = m[reading]
		} else {
			kanji[reading] = m[reading]
		}
	}

	fmt.Fprintf(w, ";; okuri-ari entries.\n")

	okuriKeys := sortKeysReverse(okuri)
	for _, key := range okuriKeys {
		value := sortValue(okuri[key])
		fmt.Fprintf(w, "%s %s\n", key, value)
	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, ";; okuri-nasi entries.\n")

	kanjiKeys := sortKeys(kanji)
	for _, key := range kanjiKeys {
		value := sortValue(kanji[key])
		fmt.Fprintf(w, "%s %s\n", key, value)
	}
}
