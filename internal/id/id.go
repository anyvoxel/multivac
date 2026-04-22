// Package id provides short, URL-safe ID generation.
package id

import (
	"crypto/rand"
	"encoding/base32"
	"io"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

var enc = base32.StdEncoding.WithPadding(base32.NoPadding)

// New returns a short, URL-safe id.
// It generates 8 random bytes and encodes them to base32 (lowercase).
//
// Typical length:
//   - without prefix: 13 chars
//   - with prefix:    len(prefix)+1+13 (e.g. "p_xxxxx..." ~ 15 chars)
func New(prefix string) (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	s := strings.ToLower(enc.EncodeToString(b))
	if prefix == "" {
		return s, nil
	}
	return prefix + "_" + s, nil
}

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
