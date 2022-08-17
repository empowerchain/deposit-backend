package organization

import (
	"context"
	"database/sql"
	"encore.app/admin"
	"encore.app/commons"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"errors"
)

type Organization struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	SigningPubKey    string `json:"signingPubKey"`
	EncryptionPubKey string `json:"encryptionPubKey"`
}

type CreateOrgParams struct {
	ID               string `json:"id" validate:"required"`
	Name             string `json:"name" validate:"required"`
	SigningPubKey    string `json:"signingPubKey" validate:"required"`
	EncryptionPubKey string `json:"encryptionPubKey" validate:"required"`
}

//encore:api auth method=POST
func CreateOrganization(ctx context.Context, params *CreateOrgParams) (*Organization, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	callingPubKey, _ := auth.UserID()
	resp, err := admin.IsAdmin(ctx, &admin.IsAdminParams{PubKey: string(callingPubKey)})
	if err != nil {
		return nil, err
	}
	if !resp.IsAdmin {
		return nil, &errs.Error{Code: errs.PermissionDenied}
	}

	_, err = sqldb.Exec(ctx, `
        INSERT INTO organization (id, name, signing_pub_key, encryption_pub_key)
        VALUES ($1, $2, $3, $4);
    `, params.ID, params.Name, params.SigningPubKey, params.EncryptionPubKey)
	if err != nil {
		return nil, err
	}

	return GetOrganization(ctx, &GetOrganizationParams{ID: params.ID})
}

type GetOrganizationParams struct {
	ID string `json:"id" validate:"required"`
}

//encore:api public method=POST
func GetOrganization(ctx context.Context, params *GetOrganizationParams) (*Organization, error) {
	var o Organization
	if err := sqldb.QueryRow(ctx, "SELECT id, name, signing_pub_key, encryption_pub_key FROM organization WHERE id=$1", params.ID).Scan(&o.ID, &o.Name, &o.SigningPubKey, &o.EncryptionPubKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &errs.Error{
				Code: errs.NotFound,
			}
		}
		return nil, err
	}

	return &o, nil
}

type AuthorizeCallerForOrgParams struct {
	OrganizationID string `json:"organizationID" validate:"required"`
}

//encore:api private method=POST
func AuthorizeCallerForOrg(ctx context.Context, params *AuthorizeCallerForOrgParams) error {
	if err := commons.Validate(params); err != nil {
		return err
	}

	org, err := GetOrganization(ctx, &GetOrganizationParams{ID: params.OrganizationID})
	if err != nil {
		return err
	}
	caller, _ := auth.UserID()

	if string(caller) == org.SigningPubKey {
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

type GetAllOrganizationsResponse struct {
	Organizations []Organization `json:"organizations"`
}

//encore:api public method=POST
func GetAllOrganizations(_ context.Context) (*GetAllOrganizationsResponse, error) {
	resp := &GetAllOrganizationsResponse{}
	rows, err := sqldb.Query(context.Background(), `
        SELECT id, name, signing_pub_key, encryption_pub_key FROM organization
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var o Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.SigningPubKey, &o.EncryptionPubKey); err != nil {
			return nil, err
		}
		resp.Organizations = append(resp.Organizations, o)
	}

	return resp, rows.Err()
}
