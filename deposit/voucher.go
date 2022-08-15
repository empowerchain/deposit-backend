package deposit

import (
	"context"
	"database/sql"
	"encore.app/admin"
	"encore.app/commons"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"errors"
	"time"
)

type Voucher struct {
	ID                  string    `json:"id"`
	VoucherDefinitionID string    `json:"voucherDefinitionID"`
	OwnerPubKey         string    `json:"ownerPubKey"`
	Invalidated         bool      `json:"invalidated"`
	CreatedAt           time.Time `json:"createdAt"`
}

// TODO: Test that voucher def gets returned everywhere
type VoucherResponse struct {
	Voucher           Voucher           `json:"voucher"`
	VoucherDefinition VoucherDefinition `json:"voucherDefinition"`
}

func mintVoucher(ctx context.Context, tx *sqldb.Tx, voucherDef *VoucherDefinition, ownerPubKey string) (string, error) {
	id := commons.GenerateID()
	if _, err := tx.Exec(ctx, `
        INSERT INTO voucher (id, voucher_definition_id, owner_pub_key, invalidated)
        VALUES ($1, $2, $3, $4);
    `, id, voucherDef.ID, ownerPubKey, false); err != nil {
		return "", err
	}

	return id, nil
}

type GetVoucherParams struct {
	VoucherID string `json:"voucherID" validate:"required"`
}

//encore:api public method=POST
func GetVoucher(ctx context.Context, params *GetVoucherParams) (*VoucherResponse, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	var v Voucher
	if err := sqldb.QueryRow(ctx, "SELECT id, voucher_definition_id, owner_pub_key, invalidated FROM voucher WHERE id=$1", params.VoucherID).Scan(&v.ID, &v.VoucherDefinitionID, &v.OwnerPubKey, &v.Invalidated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &errs.Error{
				Code: errs.NotFound,
			}
		}
		return nil, err
	}

	vd, err := GetVoucherDefinition(ctx, &GetVoucherDefinitionParams{VoucherDefinitionID: v.VoucherDefinitionID})
	if err != nil {
		return nil, err
	}

	return &VoucherResponse{
		Voucher:           v,
		VoucherDefinition: *vd,
	}, nil
}

type GetAllVouchersResponse struct {
	Vouchers []VoucherResponse `json:"vouchers"`
}

//encore:api public method=POST
func GetAllVouchers(ctx context.Context) (*GetAllVouchersResponse, error) {
	resp := &GetAllVouchersResponse{
		Vouchers: []VoucherResponse{},
	}
	rows, err := sqldb.Query(ctx, `
        SELECT id, voucher_definition_id, owner_pub_key, invalidated FROM voucher
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var v Voucher
		if err := rows.Scan(&v.ID, &v.VoucherDefinitionID, &v.OwnerPubKey, &v.Invalidated); err != nil {
			return nil, err
		}

		vd, err := GetVoucherDefinition(ctx, &GetVoucherDefinitionParams{VoucherDefinitionID: v.VoucherDefinitionID})
		if err != nil {
			return nil, err
		}

		resp.Vouchers = append(resp.Vouchers, VoucherResponse{
			Voucher:           v,
			VoucherDefinition: *vd,
		})
	}

	return resp, rows.Err()
}

type GetVouchersForUserParams struct {
	UserPubKey string `json:"userPubKey" validate:"required"`
}

type GetVouchersForUserResponse struct {
	Vouchers []VoucherResponse `json:"vouchers"`
}

//encore:api public method=POST
func GetVouchersForUser(ctx context.Context, params *GetVouchersForUserParams) (*GetVouchersForUserResponse, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	resp := &GetVouchersForUserResponse{
		Vouchers: []VoucherResponse{},
	}
	rows, err := sqldb.Query(ctx, `
        SELECT id, voucher_definition_id, owner_pub_key, invalidated, created_at FROM voucher WHERE owner_pub_key=$1 ORDER BY created_at DESC
    `, params.UserPubKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var v Voucher
		if err := rows.Scan(&v.ID, &v.VoucherDefinitionID, &v.OwnerPubKey, &v.Invalidated, &v.CreatedAt); err != nil {
			return nil, err
		}

		vd, err := GetVoucherDefinition(ctx, &GetVoucherDefinitionParams{VoucherDefinitionID: v.VoucherDefinitionID})
		if err != nil {
			return nil, err
		}
		resp.Vouchers = append(resp.Vouchers, VoucherResponse{
			Voucher:           v,
			VoucherDefinition: *vd,
		})
	}

	return resp, rows.Err()
}

type InvalidateVoucherParams struct {
	VoucherID string `json:"voucherID" validate:"required"`
}

//encore:api auth method=POST
func InvalidateVoucher(ctx context.Context, params *InvalidateVoucherParams) error {
	if err := commons.Validate(params); err != nil {
		return err
	}

	voucherRes, err := GetVoucher(ctx, &GetVoucherParams{VoucherID: params.VoucherID})
	if err != nil {
		return err
	}

	if err := authorizeCallerForVoucher(ctx, &voucherRes.Voucher); err != nil {
		return err
	}

	_, err = sqldb.Exec(ctx, "UPDATE voucher SET invalidated = true WHERE id=$1", voucherRes.Voucher.ID)
	return err
}

func authorizeCallerForVoucher(ctx context.Context, voucher *Voucher) error {
	caller, _ := auth.UserID()

	if string(caller) == voucher.OwnerPubKey {
		return nil
	}

	isAdminResp, err := admin.IsAdmin(ctx, &admin.IsAdminParams{PubKey: string(caller)})
	if err != nil {
		return err
	}

	if isAdminResp.IsAdmin {
		return nil
	}

	return &errs.Error{
		Code: errs.PermissionDenied, // TODO: Nicer authz system?
	}
}
