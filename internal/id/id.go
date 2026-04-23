// Package id provides short, URL-safe ID generation.
package id

import (
	"crypto/rand"
	"io"
	"time"

	"github.com/oklog/ulid/v2"
)

func NewULID() (string, error) {
	return NewULIDAt(time.Now())
}

func NewULIDAt(t time.Time) (string, error) {
	entropy := ulid.Monotonic(reader{r: rand.Reader}, 0)
	v, err := ulid.New(ulid.Timestamp(t.UTC()), entropy)
	if err != nil {
		return "", err
	}
	return v.String(), nil
}

type reader struct {
	r io.Reader
}

func (r reader) Read(p []byte) (int, error) {
	return io.ReadFull(r.r, p)
}
