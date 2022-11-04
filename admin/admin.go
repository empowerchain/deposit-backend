package admin

import (
	_ "embed"
	"encore.app/commons"

	"context"
	"database/sql"
	encore "encore.dev"
	"encore.dev/storage/sqldb"
	"errors"
	"log"
)

//go:embed test-fixtures.sql
var testFixtures string

type IsAdminParams struct {
	PubKey string `json:"pubKey" validate:"required"`
}

type IsAdminResponse struct {
	IsAdmin bool `json:"isAdmin"`
}

//encore:api private method=POST
func IsAdmin(ctx context.Context, params *IsAdminParams) (*IsAdminResponse, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	var id string
	if err := sqldb.QueryRow(ctx, "SELECT * FROM admin WHERE pub_key=$1", params.PubKey).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &IsAdminResponse{
				IsAdmin: false,
			}, nil
		}
		return nil, err
	}

	return &IsAdminResponse{
		IsAdmin: id != "",
	}, nil
}

//encore:api private method=POST
func InsertTestData(_ context.Context) error {
    if encore.Meta().Environment.Type == encore.EnvLocal {
		if _, err := sqldb.Exec(context.Background(), testFixtures); err != nil {
			log.Fatalln("unable to add fixtures:", err)
		}
	}

	return nil
}
