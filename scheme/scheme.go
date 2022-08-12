package scheme

import (
	"context"
	"database/sql"
	"encoding/json"
	"encore.app/commons"
	"encore.app/organization"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"errors"
	"time"
)

type Scheme struct {
	ID                string                     `json:"id"`
	Name              string                     `json:"name"`
	CreatedAt         time.Time                  `json:"createdAt"`
	CollectionPoints  []string                   `json:"collectionPoints"`
	RewardDefinitions []commons.RewardDefinition `json:"rewardDefinitions"`
	OrganizationID    string                     `json:"organizationID"`
}

type CreateSchemeParams struct {
	Name              string                     `json:"name" validate:"required"`
	OrganizationID    string                     `json:"organizationID" validate:"required"`
	RewardDefinitions []commons.RewardDefinition `json:"rewardDefinitions" validate:"required"`
}

//encore:api auth method=POST
func CreateScheme(ctx context.Context, params *CreateSchemeParams) (*Scheme, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	if err := organization.AuthorizeCallerForOrg(ctx, &organization.AuthorizeCallerForOrgParams{OrganizationID: params.OrganizationID}); err != nil {
		return nil, err
	}

	jsonb, err := json.Marshal(params.RewardDefinitions)
	if err != nil {
		return nil, err
	}

	id := commons.GenerateID()
	_, err = sqldb.Exec(ctx, `
        INSERT INTO scheme (id, organization_id, name, reward_definitions)
        VALUES ($1, $2, $3, $4);
    `, id, params.OrganizationID, params.Name, string(jsonb))
	if err != nil {
		return nil, err
	}

	return GetScheme(ctx, &GetSchemeParams{
		SchemeID: id,
	})
}

type EditSchemeParams struct {
	SchemeID          string                     `json:"schemeID"`
	RewardDefinitions []commons.RewardDefinition `json:"rewardDefinitions"`
}

//encore:api auth method=PUT
func EditScheme(ctx context.Context, params *EditSchemeParams) error {
	if err := commons.Validate(params); err != nil {
		return err
	}

	scheme, err := GetScheme(ctx, &GetSchemeParams{
		SchemeID: params.SchemeID,
	})
	if err != nil {
		return err
	}

	if err := organization.AuthorizeCallerForOrg(ctx, &organization.AuthorizeCallerForOrgParams{OrganizationID: scheme.OrganizationID}); err != nil {
		return err
	}

	if len(params.RewardDefinitions) > 0 {
		scheme.RewardDefinitions = params.RewardDefinitions
	}

	jsonb, err := json.Marshal(scheme.RewardDefinitions)
	if err != nil {
		return err
	}

	_, err = sqldb.Exec(ctx, `
        UPDATE scheme
		SET reward_definitions = $2
		WHERE id=$1
    `, scheme.ID, string(jsonb))

	return nil
}

type GetSchemeParams struct {
	SchemeID string `json:"schemeID"`
}

//encore:api public method=POST
func GetScheme(ctx context.Context, params *GetSchemeParams) (*Scheme, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	var s Scheme
	var rewardDefinitionsJson string
	if err := sqldb.QueryRow(ctx, "SELECT * FROM scheme WHERE id=$1", params.SchemeID).Scan(&s.ID, &s.OrganizationID, &s.Name, &s.CollectionPoints, &rewardDefinitionsJson, &s.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &errs.Error{
				Code: errs.NotFound,
			}
		}
		return nil, err
	}

	if err := json.Unmarshal([]byte(rewardDefinitionsJson), &s.RewardDefinitions); err != nil {
		return nil, err
	}

	return &s, nil
}

type GetAllSchemesResponse struct {
	Schemes []Scheme `json:"schemes"`
}

//encore:api public method=POST
func GetAllSchemes(_ context.Context) (resp *GetAllSchemesResponse, err error) {
	resp = &GetAllSchemesResponse{}
	rows, err := sqldb.Query(context.Background(), `
        SELECT id FROM scheme
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s Scheme
		if err := rows.Scan(&s.ID); err != nil {
			return nil, err
		}
		resp.Schemes = append(resp.Schemes, s)
	}

	return resp, rows.Err()
}

type AddCollectionPointParams struct {
	SchemeID              string `json:"schemeID" validate:"required"`
	CollectionPointPubKey string `json:"collectionPointPubKey" validate:"required"`
}

//encore:api auth method=POST
func AddCollectionPoint(ctx context.Context, params *AddCollectionPointParams) error {
	if err := commons.Validate(params); err != nil {
		return err
	}

	s, err := GetScheme(ctx, &GetSchemeParams{params.SchemeID})
	if err != nil {
		return err
	}

	if err := organization.AuthorizeCallerForOrg(ctx, &organization.AuthorizeCallerForOrgParams{OrganizationID: s.OrganizationID}); err != nil {
		return err
	}

	_, err = sqldb.Exec(ctx, "UPDATE scheme SET collection_points = array_append(collection_points, $1) WHERE id=$2", params.CollectionPointPubKey, s.ID)
	return err
}
