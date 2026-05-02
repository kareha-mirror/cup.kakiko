package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
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

		last := yomi[len(yomi)-1]
		if last >= 'a' && last <= 'z' {
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

func main() {
	kanji := map[string]string{}
	okuri := map[string]string{}
	for _, path := range os.Args[1:] {
		err := load(path, kanji, okuri)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf(";; okuri-ari entries.\n")

	okuriKeys := sortKeys(okuri)
	for _, key := range okuriKeys {
		fmt.Printf("%s %s\n", key, okuri[key])
	}

	fmt.Printf("\n")
	fmt.Printf(";; okuri-nasi entries.\n")

	kanjiKeys := sortKeys(kanji)
	for _, key := range kanjiKeys {
		fmt.Printf("%s %s\n", key, kanji[key])
	}
}
