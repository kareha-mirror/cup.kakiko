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

func NewMemDic(path string) *MemDic {
	kanji := map[string]string{}

	r, err := os.Open(path)
	if err == nil {
		_ = Load(r, kanji)
	}

	return &MemDic{
		path: path,

		kanji: kanji,
	}
}

func (d *MemDic) Save() {
	w, err := os.Create(d.path)
	if err == nil {
		Save(w, d.kanji)
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
