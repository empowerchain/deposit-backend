package deposit

import (
	"context"
	"encore.app/commons/testutils"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testSchemeID          = "testSchemeID"
	testUserID            = "testUserID"
	testCollectionPointID = "testCollectionPointID"
)

func TestRegister(t *testing.T) {
	testTable := []struct {
		params        RegisterRequest
		errorCode     errs.ErrCode
		authenticated bool
	}{
		{
			params: RegisterRequest{
				SchemeID:          testSchemeID,
				UserID:            testUserID,
				CollectionPointID: testCollectionPointID,
			},
			errorCode:     errs.OK,
			authenticated: true,
		},
		{
			params: RegisterRequest{
				SchemeID:          "doesNotExist",
				UserID:            testUserID,
				CollectionPointID: testCollectionPointID,
			},
			errorCode:     errs.NotFound,
			authenticated: true,
		},
		{
			params: RegisterRequest{
				SchemeID:          testSchemeID,
				UserID:            testUserID,
				CollectionPointID: testCollectionPointID,
			},
			errorCode:     errs.Unauthenticated,
			authenticated: false,
		},
	}

	for _, test := range testTable {
		require.NoError(t, testutils.ClearDb("deposit"))
		require.NoError(t, testutils.ClearDb("scheme"))
		require.NoError(t, insertTestData())

		var ctx context.Context
		if test.authenticated {
			ctx = testutils.GetAuthenticatedContext()
		} else {
			ctx = context.Background()
		}

		resp, err := Register(ctx, &test.params)
		_ = resp
		_ = err
		if test.errorCode == errs.OK {
			require.NoError(t, err)
			require.NotEqual(t, "", resp.DepositID)

			deposits, err := getAllDeposits()
			require.NoError(t, err)
			require.Equal(t, 1, len(deposits))
			// TODO: Check Deposit from DB is saved correctly
		} else {
			require.Error(t, err)
			require.Equal(t, test.errorCode, err.(*errs.Error).Code)

			deposits, err := getAllDeposits()
			require.NoError(t, err)
			require.Equal(t, 0, len(deposits))
		}
	}
}

func getAllDeposits() (resp []Deposit, err error) {
	rows, err := sqldb.Query(context.Background(), `
        SELECT * from deposit
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d Deposit
		if err := rows.Scan(&d.ID, &d.SchemeID, &d.CollectionPointID, &d.UserID, &d.CreatedAt); err != nil {
			return nil, err
		}
		resp = append(resp, d)
	}

	return resp, rows.Err()
}

func insertTestData() error {
	_, err := sqldb.Exec(context.Background(), `
        INSERT INTO scheme (id, name)
        VALUES ($1, $2)
    `, testSchemeID, "testSchemeName")
	return err
}
