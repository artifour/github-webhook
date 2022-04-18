package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type jsonConfigReader struct {
	schema *interface{}
}

func (reader *jsonConfigReader) load(input string, schema interface{}) error {
	command, err := os.Executable()
	if err != nil {
		panic(err)
	}

	WorkDir := filepath.Dir(command) + string(os.PathSeparator)

	file, err := os.Open(input)
	if err != nil {
		log.Fatalf("Cannot read '%s' file", WorkDir+input)
	}
	defer file.Close()

	jsonDecoder := json.NewDecoder(file)
	if jsonDecoder.Decode(&schema) != nil {
		log.Fatal("Error: ", err)
	}

	reader.schema = &schema
	return nil
}

func (reader *jsonConfigReader) get(path string) string {
	pathParts := strings.SplitN(path, ".", -1)
	high := len(pathParts) - 1
	s := *reader.schema
	for i, token := range pathParts {
		varType := reflect.TypeOf(s).String()
		if varType != "map[string]interface {}" {
			log.Fatal("Json config parse error")
		}

		if i >= high {
			return s.(map[string]interface{})[token].(string)
		}

		s = s.(map[string]interface{})[token]
	}

	return ""
}
