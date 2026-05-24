package skkdic

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type MemDic struct {
	path    string
	date    time.Time
	table   map[string]string
	loaded  bool
	removed map[string]bool
}

func NewMemDic(path string) *MemDic {
	return &MemDic{
		path:    path,
		table:   map[string]string{},
		loaded:  false,
		removed: map[string]bool{},
	}
}

func (d *MemDic) ensureLoaded() error {
	if d.loaded {
		return nil
	}

	_, err := os.Stat(d.path)
	if err != nil {
		// does not exist
		d.loaded = true
		return nil
	}

	r, err := os.Open(d.path)
	if err != nil {
		return err
	}
	err = Load(r, d.table)
	if err != nil {
		return err
	}

	d.date = time.Now()
	d.loaded = true
	return nil
}

func (d *MemDic) loadOld() (map[string]string, error) {
	info, err := os.Stat(d.path)
	if err != nil {
		return nil, err
	}
	if !info.ModTime().After(d.date) {
		return nil, fmt.Errorf("not modified")
	}

	r, err := os.Open(d.path)
	if err != nil {
		return nil, err
	}

	table := map[string]string{}
	err = Load(r, table)
	if err != nil {
		return nil, err
	}

	return table, nil
}

func (d *MemDic) Finish() error {
	if !d.loaded {
		return nil
	}

	table, err := d.loadOld()
	if err == nil {
		for reading, seq := range table {
			cands := parseSeq(seq)
			mycands, err := d.Lookup(reading)
			if err != nil {
				mycands = []string{}
			}
			for i := len(cands) - 1; i >= 0; i-- {
				cand := cands[i]
				found := false
				for _, mycand := range mycands {
					if cand == mycand {
						found = true
					}
					break
				}
				if !found {
					_, ok := d.removed[reading+" "+cand]
					if !ok {
						d.Add(reading, cand)
					}
				}
			}
		}
	}

	temp := d.path + ".temp"
	w, err := os.Create(temp)
	if err != nil {
		return err
	}
	err = Save(w, d.table)
	if err != nil {
		return err
	}
	return os.Rename(temp, d.path)
}

func (d *MemDic) Lookup(reading string) ([]string, error) {
	err := d.ensureLoaded()
	if err != nil {
		return []string{}, err
	}

	seq, ok := d.table[reading]
	if !ok {
		return []string{}, nil
	}
	cands := parseSeq(string(seq))
	return cands, nil
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

func (d *MemDic) Add(reading, word string) error {
	if reading == "" || word == "" {
		return fmt.Errorf("reading or word is null string")
	}

	err := d.ensureLoaded()
	if err != nil {
		return err
	}

	cands, err := d.Lookup(reading)
	if err != nil {
		return err
	}
	cands = removeElem(cands, word)

	buf := strings.Builder{}
	buf.WriteRune('/')
	buf.WriteString(escape(word))
	for _, cand := range cands {
		buf.WriteRune('/')
		buf.WriteString(escape(cand))
	}
	buf.WriteRune('/')

	d.table[reading] = buf.String()

	delete(d.removed, reading+" "+word)

	return nil
}

func (d *MemDic) Remove(reading, word string) error {
	err := d.ensureLoaded()
	if err != nil {
		return err
	}

	cands, err := d.Lookup(reading)
	if err != nil {
		return err
	}
	if len(cands) < 1 {
		return nil
	}

	cands = removeElem(cands, word)

	d.removed[reading+" "+word] = true

	if len(cands) < 1 {
		delete(d.table, reading)
		return nil
	}

	buf := strings.Builder{}
	for _, cand := range cands {
		buf.WriteRune('/')
		buf.WriteString(escape(cand))
	}
	buf.WriteRune('/')

	d.table[reading] = buf.String()
	return nil
}

func (d *MemDic) Sync() error {
	err := d.Finish()
	if err != nil {
		return err
	}

	d.table = map[string]string{}
	d.loaded = false
	d.removed = map[string]bool{}
	return nil
}
