package deposit

import (
	"encore.app/commons"
	"time"
)

type Deposit struct {
	ID                string
	SchemeID          string
	CollectionPointID string
	UserID            string
	CreatedAt         time.Time
}

func createDeposit(schemeID string, collectionPointID string, userID string) Deposit {
	return Deposit{
		ID:                commons.GenerateID(),
		SchemeID:          schemeID,
		CollectionPointID: collectionPointID,
		UserID:            userID,
	}
}
