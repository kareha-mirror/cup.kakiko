package skkdic

type StrDic struct {
	table  map[string]string
	loaded bool
}

func NewStrDic(data string) *StrDic {
	table := map[string]string{}
	// TODO error
	LoadStr(data, table)
	return &StrDic{
		table:  table,
		loaded: false,
	}
}

func (d *StrDic) Finish() error {
	return nil
}

func (d *StrDic) Lookup(reading string) ([]string, error) {
	seq, ok := d.table[reading]
	if !ok {
		return []string{}, nil
	}
	cands := parseSeq(string(seq))
	return cands, nil
}
