// jisyo.go - SKK-JISYO-E v1 over CDB backend
// API:
//   import "tea.kareha.org/cup/kakiko/internal/skk"
//   j := skk.NewCDBJisyo(path)
//   j.Lookup(reading)
//   j.LookupOkuri(key, okuriKana)

package skk

import (
	"github.com/colinmarc/cdb"
)

type CDBJisyo struct {
	path     string
	database *cdb.CDB
}

func NewCDBJisyo(path string) *CDBJisyo {
	return &CDBJisyo{
		path:     path,
		database: nil,
	}
}

func (j *CDBJisyo) getDb() (*cdb.CDB, error) {
	if j.database == nil {
		db, err := cdb.Open(j.path)
		if err != nil {
			return nil, err
		}
		j.database = db
	}
	return j.database, nil
}

func (j *CDBJisyo) Lookup(reading string) ([]string, error) {
	db, err := j.getDb()
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

func (j *CDBJisyo) LookupOkuri(key, okuriKana string) ([]string, error) {
	db, err := j.getDb()
	if err != nil {
		return []string{}, err
	}
	body, err := db.Get([]byte(key))
	if err != nil {
		return []string{}, err
	}
	defaults, blocks := parseBody(string(body))
	if okuriKana != "" && len(blocks) > 0 {
		result, ok := blocks[okuriKana]
		if ok && len(result) > 0 {
			return result, nil
		}
	}
	return defaults, nil
}
