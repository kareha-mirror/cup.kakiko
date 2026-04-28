package skk

import (
	"fmt"
)

type MemUserDic struct {
	kanji map[string]string
	okuri map[string]string
}

func NewMemUserDic() *MemUserDic {
	return &MemUserDic{
		kanji: make(map[string]string, 0),
		okuri: make(map[string]string, 0),
	}
}

func (d *MemUserDic) Lookup(reading string) ([]string, error) {
	body, ok := d.kanji[reading]
	if !ok {
		return []string{}, nil
	}
	defaults, _ := parseBody(string(body))
	return defaults, nil
}

func (d *MemUserDic) LookupOkuri(key, okuri string) ([]string, error) {
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

func (d *MemUserDic) Add(reading, kanji string) {
	prev, ok := d.kanji[reading]
	if ok {
		d.kanji[reading] = fmt.Sprintf("/%s%s/", kanji, prev)
	} else {
		d.kanji[reading] = fmt.Sprintf("/%s/", kanji)
	}
}

// XXX
func (d *MemUserDic) AddOkuri(key, okuri, kanji string) {
	prev, ok := d.okuri[key]
	if ok {
		d.okuri[key] = fmt.Sprintf("/%s%s", kanji, prev)
	} else {
		d.okuri[key] = fmt.Sprintf("/%s/", kanji)
	}
}
