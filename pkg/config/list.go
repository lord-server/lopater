package config

import (
	"io/ioutil"
	"github.com/hjson/hjson-go/v4"
)

func ParseList(path string) (map[string]string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data map[string]string
	err = hjson.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
