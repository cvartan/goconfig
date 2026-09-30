package filereader

import (
	"context"
	"os"

	"github.com/cvartan/goconfig/types"
)

type FileReader struct{}

func (f *FileReader) Read(ctx context.Context, source string) (fs []byte, err error) {
	fs, err = os.ReadFile(source)
	if err != nil {
		err = types.NewReadConfigurationSourceError(err, source)
	}
	return
}
