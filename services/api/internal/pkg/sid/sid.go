package sid

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

var encoder = base32.StdEncoding.WithPadding(base32.NoPadding)

func New(length int) (string, error) {
	if length < 6 {
		length = 8
	}
	n := (length*5 + 7) / 8
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	s := encoder.EncodeToString(b)
	s = strings.ToLower(s)
	if len(s) > length {
		s = s[:length]
	}
	return s, nil
}
