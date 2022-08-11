package deposit

import (
	"context"
	"database/sql"
	"encoding/json"
	"encore.app/commons"
	"encore.app/scheme"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"errors"
	"time"
)

type Deposit struct {
	ID                    string                `json:"id"`
	SchemeID              string                `json:"schemeID"`
	CollectionPointPubKey string                `json:"collectionPointPubKey"`
	UserPubKey            string                `json:"userPubKey"`
	ExternalRef           string                `json:"externalRef"`
	CreatedAt             time.Time             `json:"createdAt"`
	MassBalanceDeposits   []commons.MassBalance `json:"massBalanceDeposits"`
	Claimed               bool                  `json:"claimed"`
}

type MakeDepositParams struct {
	SchemeID            string                `json:"schemeID" validate:"required"`
	MassBalanceDeposits []commons.MassBalance `json:"massBalanceDeposits" validate:"required"`
	UserPubKey          string                `json:"userPubKey"`
	ExternalRef         string                `json:"externalRef"`
}

//encore:api auth method=POST
func MakeDeposit(ctx context.Context, params *MakeDepositParams) (*Deposit, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	collectionPoint, _ := auth.UserID()

	s, err := scheme.GetScheme(ctx, &scheme.GetSchemeParams{SchemeID: params.SchemeID})
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

	for _, deposit := range params.MassBalanceDeposits {
		depositIsAllowed := false
		for _, allowed := range s.RewardDefinitions {
			if allowed.ItemDefinition.SameAs(deposit.ItemDefinition) {
				depositIsAllowed = true
			}
		}

		if !depositIsAllowed {
			return nil, &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "no reward definition found for the deposit",
			}
		}
	}

	deposit := Deposit{
		ID:                    commons.GenerateID(),
		SchemeID:              params.SchemeID,
		CollectionPointPubKey: string(collectionPoint),
		MassBalanceDeposits:   params.MassBalanceDeposits,
		ExternalRef:           params.ExternalRef,
	}

	jsonb, err := json.Marshal(&deposit.MassBalanceDeposits)
	if err != nil {
		return nil, err
	}

	_, err = sqldb.Exec(ctx, `
	        INSERT INTO deposit (id, scheme_id, collection_point_pub_key, mass_balance_deposits, external_ref)
	        VALUES ($1, $2, $3, $4, $5)
	    `, deposit.ID, deposit.SchemeID, deposit.CollectionPointPubKey, string(jsonb), deposit.ExternalRef)
	if err != nil {
		return nil, err
	}

	if params.UserPubKey != "" {
		if _, err := Claim(ctx, &ClaimParams{
			DepositID:  deposit.ID,
			UserPubKey: params.UserPubKey,
		}); err != nil {
			// TODO: WE NEED A TRANSACTION SO THE DEPOSIT IS NOT CREATED
			return nil, err
		}
	}

	return GetDeposit(ctx, &GetDepositParams{
		DepositID: deposit.ID,
	})
}

type GetDepositParams struct {
	DepositID string `json:"depositID" validate:"required"`
}

//encore:api auth method=POST
func GetDeposit(ctx context.Context, params *GetDepositParams) (*Deposit, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	var d Deposit
	var massBalanceJson string
	if err := sqldb.QueryRow(ctx, "SELECT id, scheme_id, collection_point_pub_key, user_pub_key, mass_balance_deposits, claimed, created_at FROM deposit WHERE id=$1", params.DepositID).Scan(&d.ID, &d.SchemeID, &d.CollectionPointPubKey, &d.UserPubKey, &massBalanceJson, &d.Claimed, &d.CreatedAt); err != nil {
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

type GetDepositByExternalRefParams struct {
	CollectionPointPubKey string `json:"collectionPointPubKey" validate:"required"`
	ExternalRef           string `json:"externalRef" validate:"required"`
}

//encore:api auth method=POST
func GetDepositByExternalRef(ctx context.Context, params *GetDepositByExternalRefParams) (*Deposit, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	var d Deposit
	var massBalanceJson string
	if err := sqldb.QueryRow(ctx, "SELECT id, scheme_id, collection_point_pub_key, user_pub_key, mass_balance_deposits, claimed, created_at FROM deposit WHERE collection_point_pub_key=$1 AND external_ref=$2", params.CollectionPointPubKey, params.ExternalRef).Scan(&d.ID, &d.SchemeID, &d.CollectionPointPubKey, &d.UserPubKey, &massBalanceJson, &d.Claimed, &d.CreatedAt); err != nil {
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

type GetAllDepositsResponse struct {
	Deposits []Deposit `json:"deposits"`
}

//encore:api auth method=POST
func GetAllDeposits(_ context.Context) (*GetAllDepositsResponse, error) {
	resp := &GetAllDepositsResponse{}
	rows, err := sqldb.Query(context.Background(), `
        SELECT id from deposit
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d Deposit
		if err := rows.Scan(&d.ID); err != nil {
			return nil, err
		}
		resp.Deposits = append(resp.Deposits, d)
	}

	return resp, rows.Err()
}
