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
	ID                string
	Name              string
	CreatedAt         time.Time
	CollectionPoints  []string
	RewardDefinitions []commons.RewardDefinition
	OrganizationID    string
}

type CreateSchemeParams struct {
	Name              string `json:"name" validate:"required"`
	OrganizationID    string `json:"organizationID" validate:"required"`
	RewardDefinitions []commons.RewardDefinition
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

type GetSchemeParams struct {
	SchemeID string
}

//encore:api auth method=POST
func GetScheme(ctx context.Context, params *GetSchemeParams) (*Scheme, error) {
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

//encore:api auth method=POST
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
