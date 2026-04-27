package skk

type Jisyos struct {
	uj UserJisyo
	j  []Jisyo
}

func (j *Jisyos) Lookup(reading string) ([]string, error) {
	cands := make([]string, 0)
	if j.uj != nil {
		c, e := j.uj.Lookup(reading)
		if e == nil {
			cands = append(cands, c...)
		}
	}
	for _, jisyo := range j.j {
		c, e := jisyo.Lookup(reading)
		if e != nil {
			continue
		}
		cands = append(cands, c...)
	}
	return cands, nil
}

func (j *Jisyos) LookupOkuri(key, okuriKana string) ([]string, error) {
	cands := make([]string, 0)
	if j.uj != nil {
		c, e := j.uj.LookupOkuri(key, okuriKana)
		if e == nil {
			cands = append(cands, c...)
		}
	}
	for _, jisyo := range j.j {
		c, e := jisyo.LookupOkuri(key, okuriKana)
		if e != nil {
			continue
		}
		cands = append(cands, c...)
	}
	return cands, nil
}

func (j *Jisyos) Add(reading, gokan string) {
	j.uj.Add(reading, gokan)
}

func (j *Jisyos) AddOkuri(key, okuriKana, gokan string) {
	j.uj.AddOkuri(key, okuriKana, gokan)
}

func (j *Jisyos) SetUserJisyo(uj UserJisyo) {
	j.uj = uj
}

func (j *Jisyos) AddJisyo(jisyo Jisyo) {
	j.j = append(j.j, jisyo)
}
