package image

import (
	"errors"
	"io"
)

func Store() io.Writer {
	return &store{}
}

type store struct {
}

func (s *store) Write(p []byte) (n int, err error) {
	return 0, errors.New("not implemented")
}
