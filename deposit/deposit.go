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
	"fmt"
	"time"
)

type Deposit struct {
	ID                    string
	SchemeID              string
	CollectionPointPubKey string
	UserPubKey            string
	CreatedAt             time.Time
	MassBalanceDeposits   []commons.MassBalance
}

type MakeDepositRequest struct {
	SchemeID            string                `json:"schemeID" validate:"required"`
	MassBalanceDeposits []commons.MassBalance `json:"massBalanceDeposits" validate:"required"`
	UserPubKey          string                `json:"userPubKey"`
}

//encore:api auth method=POST
func MakeDeposit(ctx context.Context, params *MakeDepositRequest) (*Deposit, error) {
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
		fmt.Println("deposit", deposit.ItemDefinition)
		depositIsAllowed := false
		for _, allowed := range s.RewardDefinitions {
			fmt.Println("allowed", allowed)
			if allowed.ItemDefinition.SameAs(deposit.ItemDefinition) {
				depositIsAllowed = true
			}
		}

		if !depositIsAllowed {
			return nil, &errs.Error{
				Code: errs.InvalidArgument,
			}
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

type GetAllDepositsResponse struct {
	Deposits []Deposit
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

//func (s Scheme) Deposit(collectionPointPubKey string, items []MassBalance, userPubKey string) (Deposit, error) {
// If user is there, pay out reward
//}
