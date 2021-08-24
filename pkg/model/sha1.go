package model

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// Sha1 return SHA1 fingerprint
func Sha1(o interface{}) string {
	hasher := sha1.New()

	// no err check https://golang.org/pkg/hash/#Hash
	_, _ = hasher.Write([]byte(fmt.Sprintf("%#v", o)))

	return hex.EncodeToString(hasher.Sum(nil))
}
