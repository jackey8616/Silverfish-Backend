package silverfish

import (
	"crypto/sha512"
	"encoding/hex"
	"math/rand"
	"strings"
)

const dictionary string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// SHA512Str export
func SHA512Str(src, hashSalt *string) *string {
	salted := strings.Join([]string{*src, *hashSalt}, "")
	h := sha512.New()
	h.Write([]byte(salted))
	s := hex.EncodeToString(h.Sum(nil))
	return &s
}

// RandomStr export
func RandomStr(length int) *string {
	output := ""
	index := 0
	for i := 0; i < length; i++ {
		index = rand.Intn(len(dictionary))
		output += dictionary[index : index+1]
	}
	return &output
}
