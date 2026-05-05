package skkdic

type Dic interface {
	Lookup(reading string) ([]string, error)
}

type UserDic interface {
	Dic
	Add(reading, kanji string)
	Remove(reading, kanji string)
	Save()
}
