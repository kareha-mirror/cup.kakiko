package skkdic

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

func HasOkuri(reading string) bool {
	if reading == "" {
		return false
	}
	for i, r := range reading {
		if i > 0 && i == len(reading)-1 {
			if r >= 'a' && r <= 'z' { // okuri alphabet range
				return true
			}
		} else {
			if r < 0x3041 || r > 0x3096 { // normal hiragana range
				return false
			}
		}
	}
	return false
}

func LoadStr(data string, table map[string]string) error {
	lines := strings.Split(data, "\n")
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
		seq := line[space+1:]

		prev, ok := table[reading]
		if ok && len(prev) > 0 {
			table[reading] = seq + prev[1:]
		} else {
			table[reading] = seq
		}
	}

	return nil
}

func Load(r io.Reader, table map[string]string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return LoadStr(string(data), table)
}

func sortedReadings(table map[string]string) []string {
	readings := make([]string, 0, len(table))
	for reading := range table {
		readings = append(readings, reading)
	}
	sort.Strings(readings)
	return readings
}

func reverseSortedReadings(table map[string]string) []string {
	readings := make([]string, 0, len(table))
	for reading := range table {
		readings = append(readings, reading)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(readings)))
	return readings
}

func sortSeq(seq string) string {
	cands := strings.Split(seq[1:len(seq)-1], "/")
	m := map[string]bool{}
	for _, cand := range cands {
		m[cand] = true
	}
	list := []string{}
	for c := range m {
		list = append(list, c)
	}
	sort.Strings(list)
	return "/" + strings.Join(list, "/") + "/"
}

func Save(w io.Writer, table map[string]string) error {
	kanji := map[string]string{}
	okuri := map[string]string{}

	for reading := range table {
		if HasOkuri(reading) {
			okuri[reading] = table[reading]
		} else {
			kanji[reading] = table[reading]
		}
	}

	var err error

	_, err = fmt.Fprintf(w, ";; okuri-ari entries.\n")
	if err != nil {
		return err
	}

	okuriReadings := reverseSortedReadings(okuri)
	for _, reading := range okuriReadings {
		seq := sortSeq(okuri[reading])
		_, err = fmt.Fprintf(w, "%s %s\n", reading, seq)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(w, "\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, ";; okuri-nasi entries.\n")
	if err != nil {
		return err
	}

	kanjiReadings := sortedReadings(kanji)
	for _, reading := range kanjiReadings {
		seq := sortSeq(kanji[reading])
		_, err = fmt.Fprintf(w, "%s %s\n", reading, seq)
		if err != nil {
			return err
		}
	}

	return nil
}
