package scheme

import (
	"context"
	"encore.app/commons"
	"encore.dev/storage/sqldb"
)

type CreateSchemeParams struct {
	Name string `validate:"required"`
}

type CreateSchemeResponse struct {
	ID string
}

//encore:api auth method=POST
func Create(ctx context.Context, params *CreateSchemeParams) (*CreateSchemeResponse, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	scheme, err := createScheme(params.Name)
	if err != nil {
		return nil, err
	}

	if err := insert(ctx, scheme); err != nil {
		return nil, err
	}

	return &CreateSchemeResponse{ID: scheme.ID}, nil
}

func insert(ctx context.Context, scheme Scheme) error {
	_, err := sqldb.Exec(ctx, `
        INSERT INTO scheme (id, name)
        VALUES ($1, $2)
    `, scheme.ID, scheme.Name)
	return err
}

//encore:api auth method=GET
func AuthTest(_ context.Context) error {
	return nil
}

//encore:api public method=GET
func StatusTest(_ context.Context) error {
	return nil
}
