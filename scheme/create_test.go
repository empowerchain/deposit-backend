package scheme

import (
	"context"
	"encore.app/commons/testutils"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateScheme(t *testing.T) {
	testTable := []struct {
		params        CreateSchemeParams
		errorCode     errs.ErrCode
		authenticated bool
	}{
		{
			params: CreateSchemeParams{
				Name: "Valid",
			},
			errorCode:     errs.OK,
			authenticated: true,
		},
		{
			params: CreateSchemeParams{
				Name: "",
			},
			errorCode:     errs.InvalidArgument,
			authenticated: true,
		},
		{
			params: CreateSchemeParams{
				Name: "Valid",
			},
			errorCode:     errs.Unauthenticated,
			authenticated: false,
		},
	}

	for _, test := range testTable {
		require.NoError(t, testutils.ClearDb())

		var ctx context.Context
		if test.authenticated {
			ctx = testutils.GetAuthenticatedContext()
		} else {
			ctx = context.Background()
		}
		resp, err := Create(ctx, &test.params)
		if test.errorCode == errs.OK {
			require.NoError(t, err)
			require.NotEqual(t, "", resp.ID)

			schemes, err := getAllSchemes()
			require.NoError(t, err)
			require.Equal(t, 1, len(schemes))
			require.Equal(t, resp.ID, schemes[0].ID)
			require.Equal(t, test.params.Name, schemes[0].Name)
		} else {
			require.Error(t, err)
			require.Equal(t, test.errorCode, err.(*errs.Error).Code)

			schemes, err := getAllSchemes()
			require.NoError(t, err)
			require.Equal(t, 0, len(schemes))
		}
	}
}

func getAllSchemes() (resp []Scheme, err error) {
	rows, err := sqldb.Query(context.Background(), `
        SELECT * from scheme
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s Scheme
		if err := rows.Scan(&s.ID, &s.Name, &s.CreatedAt); err != nil {
			return nil, err
		}
		resp = append(resp, s)
	}

	return resp, rows.Err()
}
