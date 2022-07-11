package scheme

import (
	"context"
)

type ClaimRequest struct {
	DepositID string
}

//encore:api auth method=POST
func ClaimReward(ctx context.Context, params *ClaimRequest) error {
	return nil
}
