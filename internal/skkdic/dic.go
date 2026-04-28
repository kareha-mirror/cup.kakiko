package skkdic

type Dic interface {
	Lookup(reading string) ([]string, error)
	LookupOkuri(key, okuri string) ([]string, error)
}

type UserDic interface {
	Dic
	Add(reading, kanji string)
	AddOkuri(key, okuri, kanji string)
	Remove(reading, kanji string)
	RemoveOkuri(key, okuri, kanji string)
}
