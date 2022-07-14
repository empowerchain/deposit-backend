package commons

import "reflect"

type MagnitudeType int

const (
	Weight MagnitudeType = iota
	Count
)

type ItemDefinition struct {
	MaterialDefinition map[string]string
	Magnitude          MagnitudeType
}

func (id ItemDefinition) SameAs(diff ItemDefinition) bool {
	return reflect.DeepEqual(id, diff) && id.Magnitude == diff.Magnitude
}

type MassBalance struct {
	ItemDefinition ItemDefinition
	Amount         float64
}

type RewardType int

const (
	Token RewardType = iota
	Voucher
)

type RewardDefinition struct {
	ItemDefinition ItemDefinition
	RewardType     RewardType
	RewardPerUnit  int
}
