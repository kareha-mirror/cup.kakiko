package skk

import (
	"fmt"
)

type MemUserJisyo struct {
	gokan map[string]string
	okuri map[string]string
}

func NewMemUserJisyo() *MemUserJisyo {
	return &MemUserJisyo{
		gokan: make(map[string]string, 0),
		okuri: make(map[string]string, 0),
	}
}

func (j *MemUserJisyo) Lookup(reading string) ([]string, error) {
	body, ok := j.gokan[reading]
	if !ok {
		return []string{}, nil
	}
	defaults, _ := parseBody(string(body))
	return defaults, nil
}

func (j *MemUserJisyo) LookupOkuri(key, okuriKana string) ([]string, error) {
	body, ok := j.okuri[key]
	if !ok {
		return []string{}, nil
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

func (j *MemUserJisyo) Add(reading, gokan string) {
	prev, ok := j.gokan[reading]
	if ok {
		j.gokan[reading] = fmt.Sprintf("/%s%s/", gokan, prev)
	} else {
		j.gokan[reading] = fmt.Sprintf("/%s/", gokan)
	}
}

func (j *MemUserJisyo) AddOkuri(key, okuriKana, gokan string) {
	prev, ok := j.okuri[key]
	if ok {
		j.okuri[key] = fmt.Sprintf("/%s%s", gokan, prev)
	} else {
		j.okuri[key] = fmt.Sprintf("/%s/", gokan)
	}
}
