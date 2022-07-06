package deposit

import (
	"encore.app/commons"
	"time"
)

type Scheme struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

func createScheme(name string) Scheme {
	return Scheme{
		ID:   commons.GenerateID(),
		Name: name,
	}
}
