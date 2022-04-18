package config

var reader configReader

func Init(input string, schema interface{}) error {
	reader = createReader(input)

	return reader.load(resolvePath(input), schema)
}

func Get(path string) string {
	return reader.get(path)
}
