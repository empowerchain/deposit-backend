package deposit

import (
	"context"
	"encore.app/admin"
	"encore.app/commons"
	"encore.app/scheme"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
)

type ClaimParams struct {
	DepositID  string `json:"depositID" validate:"required"`
	UserPubKey string `json:"userPubKey" validate:"required"`
}

type ClaimResponse struct {
	Rewards []commons.Reward
}

//encore:api auth method=POST
func Claim(ctx context.Context, params *ClaimParams) (*ClaimResponse, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	deposit, err := GetDeposit(ctx, &GetDepositParams{DepositID: params.DepositID})
	if err != nil {
		return nil, err
	}

	if err := authorizeCallerToClaim(ctx, params.UserPubKey, deposit); err != nil {
		return nil, err
	}

	if deposit.Claimed {
		return nil, &errs.Error{
			Code: errs.InvalidArgument,
		}
	}

	tx, err := sqldb.Begin(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, "UPDATE deposit SET claimed = true, user_pub_key=$1 WHERE id=$2", params.UserPubKey, params.DepositID); err != nil {
		return nil, err
	}

	deposit.UserPubKey = params.UserPubKey
	rewards, err := payOutRewards(ctx, tx, deposit)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &ClaimResponse{
		Rewards: rewards,
	}, nil
}

func authorizeCallerToClaim(ctx context.Context, userPubKey string, deposit *Deposit) error {
	caller, _ := auth.UserID()

	if string(caller) == userPubKey {
		return nil
	}

	if string(caller) == deposit.CollectionPointPubKey {
		return nil
	}

	resp, err := admin.IsAdmin(ctx, &admin.IsAdminParams{PubKey: string(caller)})
	if err != nil {
		return err
	}

	if resp.IsAdmin {
		return nil
	}

	return &errs.Error{
		Code: errs.PermissionDenied,
	}
}

func payOutRewards(ctx context.Context, tx *sqldb.Tx, deposit *Deposit) ([]commons.Reward, error) {
	var rewards []commons.Reward
	s, err := scheme.GetScheme(ctx, &scheme.GetSchemeParams{SchemeID: deposit.SchemeID})
	if err != nil {
		return nil, err
	}
	for _, rd := range s.RewardDefinitions {
		for _, depItem := range deposit.MassBalanceDeposits {
			if rd.ItemDefinition.SameAs(depItem.ItemDefinition) {
				rewards = append(rewards, rd.GetRewardsFor(depItem))
			}
		}
	}

	for _, r := range rewards {
		switch typ := r.Type; typ {
		case commons.Voucher:
			if err := payOutVoucherRewards(ctx, tx, r, deposit.UserPubKey); err != nil {
				return nil, err
			}
		case commons.Token:
			panic("NOT SUPPORTED YET")
		default:
			panic("Reward type not found!")
		}
	}

	return rewards, nil
}

func payOutVoucherRewards(ctx context.Context, tx *sqldb.Tx, r commons.Reward, userPubKey string) error {
	voucherDef, err := GetVoucherDefinition(ctx, &GetVoucherDefinitionParams{VoucherDefinitionID: r.TypeID})
	if err != nil {
		return err
	}

	numberOfVouchers := int(r.Amount)
	for i := 0; i < numberOfVouchers; i++ {
		if _, err := mintVoucher(ctx, tx, voucherDef, userPubKey); err != nil {
			return err
		}
	}

	// TODO: HANDLE REMAINDERS!

	return nil
}
