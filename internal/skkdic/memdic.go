package skkdic

import (
	"fmt"
	"os"
	"strings"
)

type MemDic struct {
	path string

	kanji map[string]string
	okuri map[string]string
}

type dicRegion int

const (
	dicNone dicRegion = iota
	dicOkuri
	dicStem
)

func loadUserDic(path string) (map[string]string, map[string]string, error) {
	kanji := map[string]string{}
	okuri := map[string]string{}

	data, err := os.ReadFile(path)
	if err != nil {
		return kanji, okuri, err
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
		if region == dicOkuri {
			okuri[yomi] = cands
		} else { // dicStem
			kanji[yomi] = cands
		}
	}

	return kanji, okuri, nil
}

func NewMemDic(path string) *MemDic {
	kanji, okuri, _ := loadUserDic(path)

	return &MemDic{
		path: path,

		kanji: kanji,
		okuri: okuri,
	}
}

func (d *MemDic) Lookup(reading string) ([]string, error) {
	body, ok := d.kanji[reading]
	if !ok {
		return []string{}, nil
	}
	defaults, _ := parseBody(string(body))
	return defaults, nil
}

func (d *MemDic) LookupOkuri(key, okuri string) ([]string, error) {
	body, ok := d.okuri[key]
	if !ok {
		return []string{}, nil
	}
	defaults, blocks := parseBody(string(body))
	if okuri != "" && len(blocks) > 0 {
		result, ok := blocks[okuri]
		if ok && len(result) > 0 {
			return result, nil
		}
	}
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

// XXX
func (d *MemDic) AddOkuri(key, okuri, kanji string) {
	cands, err := d.LookupOkuri(key, okuri)
	if err != nil {
		cands = []string{}
	}
	cands = removeElem(cands, kanji)

	n := []string{kanji}
	n = append(n, cands...)
	d.okuri[key] = fmt.Sprintf("/%s/", strings.Join(n, "/"))
}

func (d *MemDic) Remove(reading, kanji string) {
	// TODO
}

// XXX
func (d *MemDic) RemoveOkuri(key, okuri, kanji string) {
	// TODO
}
