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

type VoucherDefinition struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationID"`
	Name           string `json:"name"`
	PictureURL     string `json:"pictureURL"`
}

type CreateVoucherDefinitionParams struct {
	OrganizationID string `json:"organizationID" validate:"required"`
	Name           string `json:"name" validate:"required"`
	PictureURL     string `json:"pictureURL" validate:"required"`
}

//encore:api auth method=POST
func CreateVoucherDefinition(ctx context.Context, params *CreateVoucherDefinitionParams) (*VoucherDefinition, error) {
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

//encore:api public method=POST
func GetVoucherDefinition(ctx context.Context, params *GetVoucherDefinitionParams) (*VoucherDefinition, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	var vd VoucherDefinition
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

type EditVoucherDefinitionParams struct {
	VoucherDefinitionID string `json:"voucherDefinitionID" validate:"required"`
	Name                string `json:"name" validate:"required"`
	PictureURL          string `json:"pictureURL" validate:"required"`
}

//encore:api auth method=POST
func EditVoucherDefinition(ctx context.Context, params *EditVoucherDefinitionParams) error {
	if err := commons.Validate(params); err != nil {
		return err
	}

	vd, err := GetVoucherDefinition(ctx, &GetVoucherDefinitionParams{VoucherDefinitionID: params.VoucherDefinitionID})
	if err != nil {
		return err
	}

	vd.Name = params.Name
	vd.PictureURL = params.PictureURL

	_, err = sqldb.Exec(ctx, `
        UPDATE voucher_definition
		SET name = $2, picture_url = $3
		WHERE id=$1
    `, vd.ID, vd.Name, vd.PictureURL)

	return err
}

type GetAllVoucherDefinitionsParams struct {
	OrganizationID string `json:"organizationID"`
}

type GetAllVoucherDefinitionsResponse struct {
	VoucherDefinitions []VoucherDefinition `json:"voucherDefinitions"`
}

// TODO: WRITE TESTS FOR SEARCH PARAMS
//encore:api public method=POST
func GetAllVoucherDefinitions(ctx context.Context, params *GetAllVoucherDefinitionsParams) (*GetAllVoucherDefinitionsResponse, error) {
	resp := &GetAllVoucherDefinitionsResponse{}
	var rows *sqldb.Rows
	var err error
	if params.OrganizationID == "" {
		rows, err = sqldb.Query(ctx, `SELECT id, organization_id, name, picture_url FROM voucher_definition`)
	} else {
		rows, err = sqldb.Query(ctx, `SELECT id, organization_id, name, picture_url FROM voucher_definition WHERE organization_id=$1`, params.OrganizationID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d VoucherDefinition
		if err := rows.Scan(&d.ID, &d.OrganizationID, &d.Name, &d.PictureURL); err != nil {
			return nil, err
		}
		resp.VoucherDefinitions = append(resp.VoucherDefinitions, d)
	}

	return resp, rows.Err()
}
