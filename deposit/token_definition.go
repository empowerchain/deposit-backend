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

type TokenDefinition struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationID"`
	Name           string `json:"name"`
}

type CreateTokenDefinitionParams struct {
	OrganizationID string `json:"organizationID" validate:"required"`
	Name           string `json:"name" validate:"required"`
}

//encore:api auth method=POST
func CreateTokenDefinition(ctx context.Context, params *CreateTokenDefinitionParams) (*TokenDefinition, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	if err := organization.AuthorizeCallerForOrg(ctx, &organization.AuthorizeCallerForOrgParams{OrganizationID: params.OrganizationID}); err != nil {
		return nil, err
	}

	id := commons.GenerateID()
	_, err := sqldb.Exec(ctx, `
        INSERT INTO token_definition (id, organization_id, name)
        VALUES ($1, $2, $3);
    `, id, params.OrganizationID, params.Name)
	if err != nil {
		return nil, err
	}

	return GetTokenDefinition(ctx, &GetTokenDefinitionParams{
		TokenDefinitionID: id,
	})
}

type GetTokenDefinitionParams struct {
	TokenDefinitionID string `json:"tokenDefinitionID" validate:"required"`
}

//encore:api public method=POST
func GetTokenDefinition(ctx context.Context, params *GetTokenDefinitionParams) (*TokenDefinition, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	var td TokenDefinition
	if err := sqldb.QueryRow(ctx, "SELECT id, organization_id, name FROM token_definition WHERE id=$1", params.TokenDefinitionID).Scan(&td.ID, &td.OrganizationID, &td.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &errs.Error{
				Code: errs.NotFound,
			}
		}
		return nil, err
	}

	return &td, nil
}

type EditTokenDefinitionParams struct {
	TokenDefinitionID string `json:"tokenDefinitionID" validate:"required"`
	Name              string `json:"name" validate:"required"`
}

// TODO: TEST AUTH
//encore:api auth method=POST
func EditTokenDefinition(ctx context.Context, params *EditTokenDefinitionParams) error {
	if err := commons.Validate(params); err != nil {
		return err
	}

	td, err := GetTokenDefinition(ctx, &GetTokenDefinitionParams{TokenDefinitionID: params.TokenDefinitionID})
	if err != nil {
		return err
	}

	if err := organization.AuthorizeCallerForOrg(ctx, &organization.AuthorizeCallerForOrgParams{OrganizationID: td.OrganizationID}); err != nil {
		return err
	}

	td.Name = params.Name

	_, err = sqldb.Exec(ctx, `
        UPDATE token_definition
		SET name = $2
		WHERE id=$1
    `, td.ID, td.Name)

	return err
}

type GetAllTokenDefinitionsParams struct {
	OrganizationID string `json:"organizationID"`
}

type GetAllTokenDefinitionsResponse struct {
	TokenDefinitions []TokenDefinition `json:"tokenDefinitions"`
}

// TODO: WRITE TESTS FOR SEARCH PARAMS
//encore:api public method=POST
func GetAllTokenDefinitions(ctx context.Context, params *GetAllTokenDefinitionsParams) (*GetAllTokenDefinitionsResponse, error) {
	resp := &GetAllTokenDefinitionsResponse{}
	var rows *sqldb.Rows
	var err error
	if params.OrganizationID == "" {
		rows, err = sqldb.Query(ctx, `SELECT id, organization_id, name FROM token_definition`)
	} else {
		rows, err = sqldb.Query(ctx, `SELECT id, organization_id, name FROM token_definition WHERE organization_id=$1`, params.OrganizationID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d TokenDefinition
		if err := rows.Scan(&d.ID, &d.OrganizationID, &d.Name); err != nil {
			return nil, err
		}
		resp.TokenDefinitions = append(resp.TokenDefinitions, d)
	}

	return resp, rows.Err()
}
