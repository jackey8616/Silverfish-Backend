package silverfish

import (
	"crypto/sha512"
	"encoding/hex"
	"strings"
)

// SHA512Str export
func SHA512Str(src, hashSalt *string) *string {
	salted := strings.Join([]string{*src, *hashSalt}, "")
	h := sha512.New()
	h.Write([]byte(salted))
	s := hex.EncodeToString(h.Sum(nil))
	return &s
}
