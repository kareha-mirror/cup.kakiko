package skkdic

type Dic interface {
	Finish() error
	Lookup(reading string) ([]string, error)
}

type UserDic interface {
	Dic
	Add(reading, word string) error
	Remove(reading, word string) error
}

type Dics struct {
	d    []Dic
	ud   UserDic
	diff UserDic
}

func (dics *Dics) Lookup(reading string) ([]string, error) {
	total := make([]string, 0)
	m := map[string]bool{}
	if dics.ud != nil {
		cands, e := dics.ud.Lookup(reading)
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
	for _, d := range dics.d {
		cands, e := d.Lookup(reading)
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

func (dics *Dics) Add(reading, word string) error {
	return dics.ud.Add(reading, word)
}

func (dics *Dics) Remove(reading, word string) error {
	return dics.ud.Remove(reading, word)
}

func (dics *Dics) AddDic(d Dic) {
	dics.d = append(dics.d, d)
}

func (dics *Dics) SetUserDic(ud UserDic) {
	dics.ud = ud
}

func (dics *Dics) SetDiffDic(diff UserDic) {
	dics.diff = diff
}

func (dics *Dics) AddDiff(reading, word string) error {
	return dics.diff.Add(reading, word)
}

func (dics *Dics) RemoveDiff(reading, word string) error {
	return dics.diff.Remove(reading, word)
}

func (dics *Dics) Finish() error {
	var uerr, derr error
	if dics.ud != nil {
		uerr = dics.ud.Finish()
	}
	if dics.diff != nil {
		derr = dics.diff.Finish()
	}
	if uerr != nil {
		return uerr
	}
	if derr != nil {
		return derr
	}
	return nil
}
