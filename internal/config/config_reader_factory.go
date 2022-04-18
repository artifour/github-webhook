package config

import (
	"path/filepath"
)

const (
	ExtJson = ".json"
	ExtEnv  = ".env"
)

func createReader(input string) configReader {
	ext := filepath.Ext(input)
	if ext == ExtJson {
		return &jsonConfigReader{}
	}

	return &envConfigReader{}
}
