package deposit

import (
	"context"
	"database/sql"
	"encore.app/commons"
	"encore.app/organization"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"errors"
)

type Definition struct {
	ID             string
	OrganizationID string
	Name           string
	PictureURL     string
}

type CreateVoucherDefinitionParams struct {
	OrganizationID string `json:"organizationID" validate:"required"`
	Name           string `json:"name" validate:"required"`
	PictureURL     string `json:"pictureURL" validate:"required"`
}

//encore:api auth method=POST
func CreateVoucherDefinition(ctx context.Context, params *CreateVoucherDefinitionParams) (*Definition, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	if err := organization.AuthorizeCallerForOrg(ctx, &organization.AuthorizeCallerForOrgParams{OrganizationID: params.OrganizationID}); err != nil {
		return nil, err
	}

	id := commons.GenerateID()
	_, err := sqldb.Exec(ctx, `
        INSERT INTO voucher_definition (id, organization_id, name, picture_url)
        VALUES ($1, $2, $3, $4);
    `, id, params.OrganizationID, params.Name, params.PictureURL)
	if err != nil {
		return nil, err
	}

	return GetVoucherDefinition(ctx, &GetVoucherDefinitionParams{
		VoucherDefinitionID: id,
	})
}

type GetVoucherDefinitionParams struct {
	VoucherDefinitionID string `json:"voucherDefinitionID" validate:"required"`
}

//encore:api auth method=POST
func GetVoucherDefinition(ctx context.Context, params *GetVoucherDefinitionParams) (*Definition, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	var vd Definition
	if err := sqldb.QueryRow(ctx, "SELECT id, organization_id, name, picture_url FROM voucher_definition WHERE id=$1", params.VoucherDefinitionID).Scan(&vd.ID, &vd.OrganizationID, &vd.Name, &vd.PictureURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &errs.Error{
				Code: errs.NotFound,
			}
		}
		return nil, err
	}

	return &vd, nil
}

type GetAllVoucherDefinitionsResponse struct {
	VoucherDefinitions []Definition
}

//encore:api auth method=POST
func GetAllVoucherDefinitions(_ context.Context) (*GetAllVoucherDefinitionsResponse, error) {
	resp := &GetAllVoucherDefinitionsResponse{}
	rows, err := sqldb.Query(context.Background(), `
        SELECT id, organization_id, name, picture_url FROM voucher_definition
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d Definition
		if err := rows.Scan(&d.ID, &d.OrganizationID, &d.Name, &d.PictureURL); err != nil {
			return nil, err
		}
		resp.VoucherDefinitions = append(resp.VoucherDefinitions, d)
	}

	return resp, rows.Err()
}
