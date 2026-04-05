package goconfig

type Reader interface {
	Read(source string) ([]byte, error)
}

type Parser interface {
	Parse(data []byte) (map[string]any, error)
}
