package skkdic

type Dics struct {
	ud UserDic
	d  []Dic
}

func (d *Dics) Lookup(reading string) ([]string, error) {
	cands := make([]string, 0)
	if d.ud != nil {
		c, e := d.ud.Lookup(reading)
		if e == nil {
			cands = append(cands, c...)
		}
	}
	for _, dic := range d.d {
		c, e := dic.Lookup(reading)
		if e != nil {
			continue
		}
		cands = append(cands, c...)
	}
	return cands, nil
}

func (d *Dics) LookupOkuri(key, okuri string) ([]string, error) {
	cands := make([]string, 0)
	if d.ud != nil {
		c, e := d.ud.LookupOkuri(key, okuri)
		if e == nil {
			cands = append(cands, c...)
		}
	}
	for _, dic := range d.d {
		c, e := dic.LookupOkuri(key, okuri)
		if e != nil {
			continue
		}
		cands = append(cands, c...)
	}
	return cands, nil
}

func (d *Dics) Add(reading, kanji string) {
	d.ud.Add(reading, kanji)
}

func (d *Dics) AddOkuri(key, okuri, kanji string) {
	d.ud.AddOkuri(key, okuri, kanji)
}

func (d *Dics) Remove(reading, kanji string) {
	d.ud.Remove(reading, kanji)
}

func (d *Dics) RemoveOkuri(key, okuri, kanji string) {
	d.ud.RemoveOkuri(key, okuri, kanji)
}

func (d *Dics) SetUserDic(ud UserDic) {
	d.ud = ud
}

func (d *Dics) AddDic(dic Dic) {
	d.d = append(d.d, dic)
}
