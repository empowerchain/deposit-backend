package commons

import (
	"math/big"
	"reflect"
)

type MagnitudeType int

const (
	Weight MagnitudeType = iota
	Count
)

type ItemDefinition struct {
	MaterialDefinition map[string]string `json:"materialDefinition"`
	Magnitude          MagnitudeType     `json:"magnitude"`
}

func (id ItemDefinition) SameAs(diff ItemDefinition) bool {
	return reflect.DeepEqual(id, diff) && id.Magnitude == diff.Magnitude
}

type MassBalance struct {
	ItemDefinition ItemDefinition `json:"itemDefinition"`
	Amount         float64        `json:"amount"`
}

type RewardType int

const (
	Token RewardType = iota
	Voucher
)

type RewardDefinition struct {
	ItemDefinition ItemDefinition `json:"itemDefinition"`
	RewardType     RewardType     `json:"rewardType"`
	RewardTypeID   string         `json:"rewardTypeID"`
	PerItem        float64        `json:"perItem"`
}

type Reward struct {
	Type   RewardType `json:"type"`
	TypeID string     `json:"typeID"`
	Amount float64    `json:"amount"`
}

func (rd RewardDefinition) GetRewardsFor(deposit MassBalance) Reward {
	if !deposit.ItemDefinition.SameAs(rd.ItemDefinition) {
		// This should _not_ happen. Ever. It is the caller's responsibility to make sure this can't happen.
		panic("trying to get rewards for a deposit from a reward def of the wrong type")
	}

	perItem := big.NewFloat(rd.PerItem)
	amountDeposited := big.NewFloat(deposit.Amount)

	var rewardAmount big.Float
	rewardAmount.Mul(amountDeposited, perItem)

	rewardAmountF, _ := rewardAmount.Float64()
	return Reward{
		Type:   rd.RewardType,
		TypeID: rd.RewardTypeID,
		Amount: rewardAmountF,
	}
}
