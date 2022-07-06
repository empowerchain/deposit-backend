package deposit

import (
	"context"
	"encore.app/commons"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
)

type RegisterRequest struct {
	SchemeID          string `json:"schemeID" validate:"required"`
	CollectionPointID string `json:"collectionPointID" validate:"required"`
	UserID            string `json:"userID"`
}

type RegisterResponse struct {
	DepositID string `json:"depositID"`
}

//encore:api auth method=POST
func Register(ctx context.Context, params *RegisterRequest) (*RegisterResponse, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	scheme, err := getScheme(ctx, params.SchemeID)
	if err != nil {
		return nil, err
	}
	if scheme == nil {
		return nil, &errs.Error{
			Code: errs.NotFound,
		}
	}

	deposit := createDeposit(params.SchemeID, params.CollectionPointID, params.UserID)

	if err := insertDeposit(ctx, deposit); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		DepositID: deposit.ID,
	}, nil
}

func insertDeposit(ctx context.Context, deposit Deposit) error {
	_, err := sqldb.Exec(ctx, `
        INSERT INTO deposit (id, scheme_id, collection_point_id, user_id)
        VALUES ($1, $2, $3, $4)
    `, deposit.ID, deposit.SchemeID, deposit.CollectionPointID, deposit.UserID)
	return err
}
