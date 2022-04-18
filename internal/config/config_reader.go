package config

type configReader interface {
	load(input string, schema interface{}) error
	get(path string) string
}
