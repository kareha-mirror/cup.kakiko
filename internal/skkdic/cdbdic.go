package skkdic

import (
	"github.com/colinmarc/cdb"
)

type CDBDic struct {
	path string
	db   *cdb.CDB
}

func NewCDBDic(path string) *CDBDic {
	return &CDBDic{
		path: path,
		db:   nil,
	}
}

func (d *CDBDic) Finish(save bool) error {
	if d.db == nil {
		return nil
	}
	return d.db.Close()
}

func (d *CDBDic) getDb() (*cdb.CDB, error) {
	if d.db == nil {
		db, err := cdb.Open(d.path)
		if err != nil {
			return nil, err
		}
		d.db = db
	}
	return d.db, nil
}

func (d *CDBDic) Lookup(reading string) ([]string, error) {
	db, err := d.getDb()
	if err != nil {
		return []string{}, err
	}
	seq, err := db.Get([]byte(reading))
	if err != nil {
		return []string{}, err
	}
	cands := parseSeq(string(seq))
	return cands, nil
}
