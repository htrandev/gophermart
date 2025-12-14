package handler

import (
	"fmt"
	"io"
)

var _ io.ReadCloser = (*ErrorReader)(nil)

type ErrorReader struct{}

func (e *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("test read error")
}

func (e *ErrorReader) Close() error {
	return nil
}
