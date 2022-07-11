package scheme

import (
	"time"
)

type MagnitudeType int

const (
	Weight MagnitudeType = iota
	Count
)

type Deposit struct {
	ID                    string
	SchemeID              string
	CollectionPointPubKey string
	UserPubKey            string
	CreatedAt             time.Time
	MassBalanceDeposits   []MassBalance
}

type ItemDefinition struct {
	MaterialDefinition map[string]string
	Magnitude          MagnitudeType
}

type MassBalance struct {
	ItemDefinition ItemDefinition
	Amount         float64
}
