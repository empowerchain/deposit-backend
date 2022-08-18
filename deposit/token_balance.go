package deposit

type TokenBalance struct {
	TokenDefinitionID string  `json:"tokenDefinitionID"`
	OwnerPubKey       string  `json:"ownerPubKey"`
	Balance           float64 `json:"balance"`
}

type GetTokenBalanceParams struct {
	TokenDefinitionID string `json:"tokenDefinitionID" validate:"required"`
	OwnerPubKey       string `json:"ownerPubKey" validate:"required"`
}

//func