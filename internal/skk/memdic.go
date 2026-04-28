package skk

import (
	"fmt"
	"strings"
)

type MemDic struct {
	kanji map[string]string
	okuri map[string]string
}

func NewMemDic() *MemDic {
	return &MemDic{
		kanji: make(map[string]string, 0),
		okuri: make(map[string]string, 0),
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
