package types

import "context"

type Reader interface {
	Read(ctx context.Context, source string) ([]byte, error)
}
