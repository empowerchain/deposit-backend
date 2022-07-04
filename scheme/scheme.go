package scheme

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

type Scheme struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

func createScheme(name string) (Scheme, error) {
	id, err := generateID()
	if err != nil {
		return Scheme{}, err
	}

	return Scheme{
		ID:   id,
		Name: name,
	}, nil
}

func generateID() (string, error) {
	var data [6]byte // 6 bytes of entropy
	if _, err := rand.Read(data[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data[:]), nil
}
