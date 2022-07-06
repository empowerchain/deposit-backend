package deposit

import (
	"context"
	"database/sql"
	"encore.app/commons"
	"encore.dev/storage/sqldb"
	"errors"
)

type CreateSchemeParams struct {
	Name string `validate:"required"`
}

type CreateSchemeResponse struct {
	ID string
}

//encore:api auth method=POST
func CreateScheme(ctx context.Context, params *CreateSchemeParams) (*CreateSchemeResponse, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	scheme := createScheme(params.Name)

	if err := insertScheme(ctx, scheme); err != nil {
		return nil, err
	}

	return &CreateSchemeResponse{ID: scheme.ID}, nil
}

func insertScheme(ctx context.Context, scheme Scheme) error {
	_, err := sqldb.Exec(ctx, `
        INSERT INTO scheme (id, name)
        VALUES ($1, $2);
    `, scheme.ID, scheme.Name)
	return err
}

func getScheme(ctx context.Context, id string) (*Scheme, error) {
	var s Scheme
	if err := sqldb.QueryRow(ctx, "SELECT * FROM scheme WHERE id=$1", id).Scan(&s.ID, &s.Name, &s.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}

//encore:api auth method=GET
func AuthTest(_ context.Context) error {
	return nil
}

//encore:api public method=GET
func StatusTest(_ context.Context) error {
	return nil
}
