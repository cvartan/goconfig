package filereader

import "os"

type FileReader struct{}

func (f *FileReader) Read(source string) (fs []byte, err error) {
	fs, err = os.ReadFile(source)
	return
}
