package commons

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateID() string {
	var data [6]byte // 6 bytes of entropy
	if _, err := rand.Read(data[:]); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data[:])
}
