package deposit

import (
	"context"
)

type ClaimRequest struct {
	DepositID string
}

//encore:api auth method=POST
func Claim(ctx context.Context, params *ClaimRequest) error {
	return nil
}
