package skkdic

import (
	"fmt"
	"os"
	"strings"
)

type MemDic struct {
	path string

	kanji map[string]string
}

type dicRegion int

const (
	dicNone dicRegion = iota
	dicOkuri
	dicStem
)

func loadUserDic(path string) (map[string]string, error) {
	kanji := map[string]string{}

	data, err := os.ReadFile(path)
	if err != nil {
		return kanji, err
	}

	lines := strings.Split(string(data), "\n")
	region := dicNone
	for _, line := range lines {
		if strings.HasPrefix(line, ";; okuri-ari entries.") {
			region = dicOkuri
			continue
		}
		if strings.HasPrefix(line, ";; okuri-nasi entries.") {
			region = dicStem
			continue
		}
		if region == dicNone {
			continue
		}
		if strings.HasPrefix(line, ";") {
			continue
		}

		space := strings.Index(line, " ")
		if space < 0 {
			continue
		}
		yomi := line[:space]
		cands := line[space+1:]
		kanji[yomi] = cands
	}

	return kanji, nil
}

func NewMemDic(path string) *MemDic {
	kanji, _ := loadUserDic(path)

	return &MemDic{
		path: path,

		kanji: kanji,
	}
}

func (d *MemDic) Lookup(reading string) ([]string, error) {
	body, ok := d.kanji[reading]
	if !ok {
		return []string{}, nil
	}
	defaults := parseBody(string(body))
	return defaults, nil
}

func removeElem(list []string, elem string) []string {
	for i, s := range list {
		if s == elem {
			n := []string{}
			n = append(n, list[:i]...)
			if i+1 < len(list) {
				n = append(n, list[i+1:]...)
			}
			return n
		}
	}
	return list
}

func (d *MemDic) Add(reading, kanji string) {
	cands, err := d.Lookup(reading)
	if err != nil {
		cands = []string{}
	}
	cands = removeElem(cands, kanji)

	n := []string{kanji}
	n = append(n, cands...)
	d.kanji[reading] = fmt.Sprintf("/%s/", strings.Join(n, "/"))
}

func (d *MemDic) Remove(reading, kanji string) {
	// TODO
}
