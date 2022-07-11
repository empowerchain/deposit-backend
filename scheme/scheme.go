package scheme

import (
	"context"
	"database/sql"
	"encoding/json"
	"encore.app/commons"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"errors"
	"fmt"
	"time"
)

type Scheme struct {
	ID               string
	Name             string
	CreatedAt        time.Time
	CollectionPoints []string
	AllowedItems     []ItemDefinition
}

type CreateSchemeParams struct {
	Name         string `validate:"required"`
	AllowedItems []ItemDefinition
}

//encore:api auth method=POST
func CreateScheme(ctx context.Context, params *CreateSchemeParams) (*Scheme, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	scheme := Scheme{
		ID:   commons.GenerateID(),
		Name: params.Name,
	}

	_, err := sqldb.Exec(ctx, `
        INSERT INTO scheme (id, name)
        VALUES ($1, $2);
    `, scheme.ID, scheme.Name)
	if err != nil {
		return nil, err
	}

	return GetScheme(ctx, &GetSchemeParams{
		SchemeID: scheme.ID,
	})
}

type GetSchemeParams struct {
	SchemeID string
}

//encore:api auth method=POST
func GetScheme(ctx context.Context, params *GetSchemeParams) (*Scheme, error) {
	var s Scheme
	if err := sqldb.QueryRow(ctx, "SELECT * FROM scheme WHERE id=$1", params.SchemeID).Scan(&s.ID, &s.Name, &s.CollectionPoints, &s.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &errs.Error{
				Code: errs.NotFound,
			}
		}
		return nil, err
	}

	return &s, nil
}

type AddCollectionPointParams struct {
	SchemeID              string
	CollectionPointPubKey string
}

//encore:api auth method=POST
func AddCollectionPoint(ctx context.Context, params *AddCollectionPointParams) error {
	s, err := GetScheme(ctx, &GetSchemeParams{params.SchemeID})
	if err != nil {
		return err
	}

	_, err = sqldb.Exec(ctx, "UPDATE scheme SET collection_points = array_append(collection_points, $1) WHERE id=$2", params.CollectionPointPubKey, s.ID)
	return err
}

type DepositRequest struct {
	SchemeID            string        `json:"schemeID" validate:"required"`
	MassBalanceDeposits []MassBalance `json:"massBalanceDeposits" validate:"required"`
	UserPubKey          string        `json:"userPubKey"`
}

//encore:api auth method=POST
func MakeDeposit(ctx context.Context, params *DepositRequest) (*Deposit, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	collectionPoint, _ := auth.UserID()

	s, err := GetScheme(ctx, &GetSchemeParams{params.SchemeID})
	if err != nil {
		return nil, err
	}

	collectionPointAllowed := false
	for _, c := range s.CollectionPoints {
		if c == string(collectionPoint) {
			collectionPointAllowed = true
			break
		}
	}
	if !collectionPointAllowed {
		return nil, &errs.Error{
			Code: errs.PermissionDenied,
		}
	}

	deposit := Deposit{
		ID:                    commons.GenerateID(),
		SchemeID:              params.SchemeID,
		CollectionPointPubKey: string(collectionPoint),
		UserPubKey:            params.UserPubKey,
		MassBalanceDeposits:   params.MassBalanceDeposits,
	}

	jsonb, err := json.Marshal(&deposit.MassBalanceDeposits)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(jsonb))

	_, err = sqldb.Exec(ctx, `
	        INSERT INTO deposit (id, scheme_id, collection_point_pub_key, user_pub_key, mass_balance_deposits)
	        VALUES ($1, $2, $3, $4, $5)
	    `, deposit.ID, deposit.SchemeID, deposit.CollectionPointPubKey, deposit.UserPubKey, string(jsonb))
	if err != nil {
		return nil, err
	}

	return GetDeposit(ctx, &GetDepositRequest{
		DepositID: deposit.ID,
	})
}

type GetDepositRequest struct {
	DepositID string
}

//encore:api auth method=POST
func GetDeposit(ctx context.Context, params *GetDepositRequest) (*Deposit, error) {
	var d Deposit
	var massBalanceJson string
	if err := sqldb.QueryRow(ctx, "SELECT * FROM deposit WHERE id=$1", params.DepositID).Scan(&d.ID, &d.SchemeID, &d.CollectionPointPubKey, &d.UserPubKey, &massBalanceJson, &d.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &errs.Error{
				Code: errs.NotFound,
			}
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(massBalanceJson), &d.MassBalanceDeposits); err != nil {
		return nil, err
	}
	return &d, nil
}

//func (s Scheme) Deposit(collectionPointPubKey string, items []MassBalance, userPubKey string) (Deposit, error) {
// Check items are OK to deposit in this scheme

// If user is there, pay out reward
//}
