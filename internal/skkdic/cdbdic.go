// cdbdic.go - SKK-JISYO-E v1 over CDB backend
// API:
//   import "tea.kareha.org/cup/kakiko/internal/skkdic"
//   d := skkdic.NewCDBDic(path)
//   d.Lookup(reading)
//   d.LookupOkuri(key, okuri)

package skkdic

import (
	"github.com/colinmarc/cdb"
)

type CDBDic struct {
	path     string
	database *cdb.CDB
}

func NewCDBDic(path string) *CDBDic {
	return &CDBDic{
		path:     path,
		database: nil,
	}
}

func (d *CDBDic) getDb() (*cdb.CDB, error) {
	if d.database == nil {
		db, err := cdb.Open(d.path)
		if err != nil {
			return nil, err
		}
		d.database = db
	}
	return d.database, nil
}

func (d *CDBDic) Lookup(reading string) ([]string, error) {
	db, err := d.getDb()
	if err != nil {
		return []string{}, err
	}
	body, err := db.Get([]byte(reading))
	if err != nil {
		return []string{}, err
	}
	defaults, _ := parseBody(string(body))
	return defaults, nil
}

func (d *CDBDic) LookupOkuri(key, okuri string) ([]string, error) {
	db, err := d.getDb()
	if err != nil {
		return []string{}, err
	}
	body, err := db.Get([]byte(key))
	if err != nil {
		return []string{}, err
	}
	defaults, blocks := parseBody(string(body))
	if okuri != "" && len(blocks) > 0 {
		result, ok := blocks[okuri]
		if ok && len(result) > 0 {
			return result, nil
		}
	}
	return defaults, nil
}
