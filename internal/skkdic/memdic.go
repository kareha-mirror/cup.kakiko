package skkdic

import (
	"errors"
	"os"
	"strings"
)

type MemDic struct {
	path   string
	table  map[string]string
	loaded bool
}

func NewMemDic(path string) *MemDic {
	return &MemDic{
		path:   path,
		table:  map[string]string{},
		loaded: false,
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

	d.loaded = true
	return nil
}

func (d *MemDic) Finish() error {
	if !d.loaded {
		return nil
	}

	w, err := os.Create(d.path)
	if err != nil {
		return err
	}
	return Save(w, d.table)
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
		return errors.New("reading or word is null string")
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
