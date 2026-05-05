package skkdic

type Dics struct {
	ud UserDic
	d  []Dic
}

func (d *Dics) Lookup(reading string) ([]string, error) {
	total := make([]string, 0)
	m := map[string]bool{}
	if d.ud != nil {
		cands, e := d.ud.Lookup(reading)
		if e == nil {
			for _, c := range cands {
				_, ok := m[c]
				if ok {
					continue
				}
				total = append(total, c)
				m[c] = true
			}
		}
	}
	for _, dic := range d.d {
		cands, e := dic.Lookup(reading)
		if e == nil {
			for _, c := range cands {
				_, ok := m[c]
				if ok {
					continue
				}
				total = append(total, c)
				m[c] = true
			}
		}
	}
	return total, nil
}

func (d *Dics) Add(reading, kanji string) {
	d.ud.Add(reading, kanji)
}

func (d *Dics) Remove(reading, kanji string) {
	d.ud.Remove(reading, kanji)
}

func (d *Dics) SetUserDic(ud UserDic) {
	d.ud = ud
}

func (d *Dics) AddDic(dic Dic) {
	d.d = append(d.d, dic)
}

func (d *Dics) Save() {
	d.ud.Save()
}
