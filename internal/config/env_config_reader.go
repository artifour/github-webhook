package config

import (
	"os"
	"strings"
)

type envConfigReader struct{}

func (*envConfigReader) load(input string, schema interface{}) error {
	return nil
}

func (*envConfigReader) get(path string) string {
	return os.Getenv(strings.ReplaceAll(path, ".", "_"))
}
