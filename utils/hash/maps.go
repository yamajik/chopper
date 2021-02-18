package hash

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// Encode bulabula
// Copied from https://github.com/kubernetes/kubernetes
// /blob/master/pkg/kubectl/util/hash/hash.go
func Encode(hex string) (string, error) {
	if len(hex) < 10 {
		return "", errors.New("input length must be at least 10")
	}
	enc := []rune(hex[:10])
	for i := range enc {
		switch enc[i] {
		case '0':
			enc[i] = 'g'
		case '1':
			enc[i] = 'h'
		case '3':
			enc[i] = 'k'
		case 'a':
			enc[i] = 'm'
		case 'e':
			enc[i] = 't'
		}
	}
	return string(enc), nil
}

// SHA256 returns the hex form of the sha256 of the argument.
func SHA256(data string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}

// FromString bulabula
func FromString(data string) (string, error) {
	return Encode(SHA256(data))
}

// FromMap bulabula
func FromMap(m map[string]interface{}) (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return FromString(string(data))
}
