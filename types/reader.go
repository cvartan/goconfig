package types

type Reader interface {
	Read(source string) ([]byte, error)
}
