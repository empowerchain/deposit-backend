package scheme

import (
	"context"
	"encore.app/commons/testutils"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testUserPubKey = "testUserID"
)

var defaultTestDeposit = []MassBalance{
	{
		ItemDefinition: ItemDefinition{
			MaterialDefinition: map[string]string{"materialType": "PET"},
			Magnitude:          Weight,
		},
		Amount: 12,
	},
}

func TestCreateScheme(t *testing.T) {
	testTable := []struct {
		name          string
		params        CreateSchemeParams
		errorCode     errs.ErrCode
		authenticated bool
	}{
		{
			name: "Happy path",
			params: CreateSchemeParams{
				Name: "Valid",
			},
			errorCode:     errs.OK,
			authenticated: true,
		},
		{
			name: "Invalid: empty name",
			params: CreateSchemeParams{
				Name: "",
			},
			errorCode:     errs.InvalidArgument,
			authenticated: true,
		},
		{
			name: "Unauthenticated",
			params: CreateSchemeParams{
				Name: "Valid",
			},
			errorCode:     errs.Unauthenticated,
			authenticated: false,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, testutils.ClearDb("scheme"))

			var ctx context.Context
			if test.authenticated {
				ctx = testutils.GetAuthenticatedContext("")
			} else {
				ctx = context.Background()
			}
			resp, err := CreateScheme(ctx, &test.params)
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
		})
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
		if err := rows.Scan(&s.ID, &s.Name, &s.CollectionPoints, &s.CreatedAt); err != nil {
			return nil, err
		}
		resp = append(resp, s)
	}

	return resp, rows.Err()
}

func TestMakeDeposit(t *testing.T) {
	collectionPointPubKey, _ := testutils.GenerateKeys()
	notCollectionPointPubKey, _ := testutils.GenerateKeys()
	testScheme, err := CreateScheme(testutils.GetAuthenticatedContext(""), &CreateSchemeParams{
		Name:         "TestScheme",
		AllowedItems: []ItemDefinition{},
	})
	require.NoError(t, err)

	err = AddCollectionPoint(testutils.GetAuthenticatedContext(""), &AddCollectionPointParams{
		SchemeID:              testScheme.ID,
		CollectionPointPubKey: collectionPointPubKey,
	})
	require.NoError(t, err)

	testTable := []struct {
		name          string
		params        DepositRequest
		errorCode     errs.ErrCode
		authenticated bool
		uid           string
	}{
		{
			name: "Happy path",
			params: DepositRequest{
				SchemeID:            testScheme.ID,
				UserPubKey:          testUserPubKey,
				MassBalanceDeposits: defaultTestDeposit,
			},
			errorCode:     errs.OK,
			authenticated: true,
			uid:           collectionPointPubKey,
		},
		{
			name: "Missing items",
			params: DepositRequest{
				SchemeID:   testScheme.ID,
				UserPubKey: testUserPubKey,
			},
			errorCode:     errs.InvalidArgument,
			authenticated: true,
			uid:           collectionPointPubKey,
		},
		{
			name: "Scheme not found",
			params: DepositRequest{
				SchemeID:            "doesNotExist",
				UserPubKey:          testUserPubKey,
				MassBalanceDeposits: defaultTestDeposit,
			},
			errorCode:     errs.NotFound,
			authenticated: true,
			uid:           collectionPointPubKey,
		},
		{
			name: "Unauthenticated",
			params: DepositRequest{
				SchemeID:            testScheme.ID,
				UserPubKey:          testUserPubKey,
				MassBalanceDeposits: defaultTestDeposit,
			},
			errorCode:     errs.Unauthenticated,
			authenticated: false,
		},
		{
			name: "Unauthorized collection point",
			params: DepositRequest{
				SchemeID:            testScheme.ID,
				UserPubKey:          testUserPubKey,
				MassBalanceDeposits: defaultTestDeposit,
			},
			errorCode:     errs.PermissionDenied,
			authenticated: true,
			uid:           notCollectionPointPubKey,
		},
		{
			name: "Item not allowed",
			params: DepositRequest{
				SchemeID:   testScheme.ID,
				UserPubKey: testUserPubKey,
				MassBalanceDeposits: []MassBalance{
					{
						ItemDefinition: ItemDefinition{
							MaterialDefinition: map[string]string{"materialType": "NOTPET"},
							Magnitude:          Weight,
						},
						Amount: 32,
					},
				},
			},
			errorCode:     errs.InvalidArgument,
			authenticated: true,
			uid:           collectionPointPubKey,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, testutils.ClearDb("deposit"))

			var ctx context.Context
			if test.authenticated {
				ctx = testutils.GetAuthenticatedContext(test.uid)
			} else {
				ctx = context.Background()
			}

			deposit, err := MakeDeposit(ctx, &test.params)
			if test.errorCode == errs.OK {
				require.NoError(t, err)
				require.NotEqual(t, "", deposit.ID)

				all, err := getAllDeposits()
				require.NoError(t, err)
				require.Equal(t, 1, len(all))

				dbDeposit, err := GetDeposit(ctx, &GetDepositRequest{
					DepositID: deposit.ID,
				})
				require.Equal(t, test.params.SchemeID, dbDeposit.SchemeID)
				require.Equal(t, test.params.UserPubKey, dbDeposit.UserPubKey)
				require.Equal(t, test.params.MassBalanceDeposits, dbDeposit.MassBalanceDeposits)
			} else {
				require.Error(t, err)
				require.Equal(t, test.errorCode, err.(*errs.Error).Code)

				deposits, err := getAllDeposits()
				require.NoError(t, err)
				require.Equal(t, 0, len(deposits))
			}
		})
	}
}

func getAllDeposits() (resp []Deposit, err error) {
	rows, err := sqldb.Query(context.Background(), `
        SELECT id from deposit
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d Deposit
		if err := rows.Scan(&d.ID); err != nil {
			return nil, err
		}
		resp = append(resp, d)
	}

	return resp, rows.Err()
}
