package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const dictionaryURL = "https://tea.kareha.org/ja/skk-e/raw/branch/main/legacy-cdb/skk-edic-legacy-large.cdb"
const dictionaryFilename = "skk-edic-legacy-large.cdb"

func getHTTP(u *url.URL) ([]byte, error) {
	resp, err := http.Get(u.String())
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return body, nil
}

func downloadDictionary(dir string) error {
	path := filepath.Join(dir, dictionaryFilename)
	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return err
	}

	u, err := url.Parse(dictionaryURL)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading dictionary..\n")
	body, err := getHTTP(u)

	err = os.WriteFile(path, body, 0666)
	if err != nil {
		return err
	}
	fmt.Printf("Done.\n")

	return nil
}
