package deposit

import (
	"context"
	"encore.app/admin"
	"encore.app/commons"
	"encore.app/commons/testutils"
	"encore.app/organization"
	"encore.app/scheme"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testUserPubKey     = "testUserID"
	testOrganizationId = "myOrgId1"
)

var (
	depositDB          = sqldb.Named("deposit")
	defaultTestRewards = commons.RewardDefinition{
		ItemDefinition: commons.ItemDefinition{
			MaterialDefinition: map[string]string{"materialType": "PET"},
			Magnitude:          commons.Weight,
		},
		RewardType:   commons.Voucher,
		RewardTypeID: "",
		PerItem:      1,
	}
	defaultTestDeposit = []commons.MassBalance{
		{
			ItemDefinition: defaultTestRewards.ItemDefinition,
			Amount:         12,
		},
	}
)

func TestMakeDeposit(t *testing.T) {
	require.NoError(t, admin.InsertTestData(context.Background()))
	testutils.ClearAllDBs()

	organizationPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:     testOrganizationId,
		Name:   testOrganizationId,
		PubKey: organizationPubKey,
	})
	require.NoError(t, err)

	definition, err := CreateVoucherDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &CreateVoucherDefinitionParams{
		OrganizationID: testOrganizationId,
		Name:           "Voucher def name",
		PictureURL:     "https://does.not.matter.com",
	})
	require.NoError(t, err)
	defaultTestRewards.RewardTypeID = definition.ID

	collectionPointPubKey, _ := testutils.GenerateKeys()
	notCollectionPointPubKey, _ := testutils.GenerateKeys()
	testScheme, err := scheme.CreateScheme(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &scheme.CreateSchemeParams{
		Name: "TestScheme",
		RewardDefinitions: []commons.RewardDefinition{
			defaultTestRewards,
		},
		OrganizationID: testOrganizationId,
	})
	require.NoError(t, err)

	err = scheme.AddCollectionPoint(testutils.GetAuthenticatedContext(organizationPubKey), &scheme.AddCollectionPointParams{
		SchemeID:              testScheme.ID,
		CollectionPointPubKey: collectionPointPubKey,
	})
	require.NoError(t, err)

	testTable := []struct {
		name      string
		params    MakeDepositParams
		errorCode errs.ErrCode
		uid       string
	}{
		{
			name: "Happy path with user",
			params: MakeDepositParams{
				SchemeID:            testScheme.ID,
				UserPubKey:          testUserPubKey,
				MassBalanceDeposits: defaultTestDeposit,
			},
			errorCode: errs.OK,
			uid:       collectionPointPubKey,
		},
		{
			name: "Missing items",
			params: MakeDepositParams{
				SchemeID:   testScheme.ID,
				UserPubKey: testUserPubKey,
			},
			errorCode: errs.InvalidArgument,
			uid:       collectionPointPubKey,
		},
		{
			name: "Scheme not found",
			params: MakeDepositParams{
				SchemeID:            "doesNotExist",
				UserPubKey:          testUserPubKey,
				MassBalanceDeposits: defaultTestDeposit,
			},
			errorCode: errs.NotFound,
			uid:       collectionPointPubKey,
		},
		{
			name: "Unauthenticated",
			params: MakeDepositParams{
				SchemeID:            testScheme.ID,
				UserPubKey:          testUserPubKey,
				MassBalanceDeposits: defaultTestDeposit,
			},
			errorCode: errs.PermissionDenied,
		},
		{
			name: "Unauthorized collection point",
			params: MakeDepositParams{
				SchemeID:            testScheme.ID,
				UserPubKey:          testUserPubKey,
				MassBalanceDeposits: defaultTestDeposit,
			},
			errorCode: errs.PermissionDenied,
			uid:       notCollectionPointPubKey,
		},
		{
			name: "Item not allowed",
			params: MakeDepositParams{
				SchemeID:   testScheme.ID,
				UserPubKey: testUserPubKey,
				MassBalanceDeposits: []commons.MassBalance{
					{
						ItemDefinition: commons.ItemDefinition{
							MaterialDefinition: map[string]string{"materialType": "NOTPET"},
							Magnitude:          commons.Weight,
						},
						Amount: 32,
					},
				},
			},
			errorCode: errs.InvalidArgument,
			uid:       collectionPointPubKey,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			ctx := testutils.GetAuthenticatedContext(test.uid)

			deposit, err := MakeDeposit(ctx, &test.params)
			if test.errorCode == errs.OK {
				require.NoError(t, err)
				require.NotEqual(t, "", deposit.ID)

				getAllResp, err := GetAllDeposits(testutils.GetAuthenticatedContext(""))
				require.NoError(t, err)
				require.Equal(t, 1, len(getAllResp.Deposits))

				dbDeposit, err := GetDeposit(ctx, &GetDepositParams{
					DepositID: deposit.ID,
				})
				require.NoError(t, err)
				require.Equal(t, test.params.SchemeID, dbDeposit.SchemeID)
				require.Equal(t, test.params.UserPubKey, dbDeposit.UserPubKey)
				require.Equal(t, test.params.MassBalanceDeposits, dbDeposit.MassBalanceDeposits)
				shouldBeClaimed := dbDeposit.UserPubKey != ""
				require.Equal(t, shouldBeClaimed, dbDeposit.Claimed)
			} else {
				require.Error(t, err)
				require.Equal(t, test.errorCode, err.(*errs.Error).Code)

				getAllResp, err := GetAllDeposits(testutils.GetAuthenticatedContext(""))
				require.NoError(t, err)
				require.Equal(t, 0, len(getAllResp.Deposits))
			}

			require.NoError(t, testutils.ClearDB(depositDB, "deposit"))
		})
	}
}

func TestGetDeposit(t *testing.T) {
	require.NoError(t, admin.InsertTestData(context.Background()))
	testutils.ClearAllDBs()

	organizationPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:     testOrganizationId,
		Name:   testOrganizationId,
		PubKey: organizationPubKey,
	})
	require.NoError(t, err)

	definition, err := CreateVoucherDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &CreateVoucherDefinitionParams{
		OrganizationID: testOrganizationId,
		Name:           "Voucher def name",
		PictureURL:     "https://does.not.matter.com",
	})
	require.NoError(t, err)
	defaultTestRewards.RewardTypeID = definition.ID

	collectionPointPubKey, _ := testutils.GenerateKeys()
	testScheme, err := scheme.CreateScheme(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &scheme.CreateSchemeParams{
		Name: "TestScheme",
		RewardDefinitions: []commons.RewardDefinition{
			defaultTestRewards,
		},
		OrganizationID: testOrganizationId,
	})
	require.NoError(t, err)

	err = scheme.AddCollectionPoint(testutils.GetAuthenticatedContext(organizationPubKey), &scheme.AddCollectionPointParams{
		SchemeID:              testScheme.ID,
		CollectionPointPubKey: collectionPointPubKey,
	})
	require.NoError(t, err)

	externalRef := "myExternalRef123"
	ctx := testutils.GetAuthenticatedContext(collectionPointPubKey)
	deposit, err := MakeDeposit(ctx, &MakeDepositParams{
		SchemeID:            testScheme.ID,
		UserPubKey:          testUserPubKey,
		MassBalanceDeposits: defaultTestDeposit,
		ExternalRef:         externalRef,
	})
	require.NoError(t, err)

	getDepositWithId, err := GetDeposit(ctx, &GetDepositParams{
		DepositID: deposit.ID,
	})
	require.NoError(t, err)
	require.Equal(t, deposit.ID, getDepositWithId.ID)

	getDepositWithExternalRef, err := GetDepositByExternalRef(ctx, &GetDepositByExternalRefParams{
		CollectionPointPubKey: collectionPointPubKey,
		ExternalRef:           externalRef,
	})
	require.NoError(t, err)
	require.Equal(t, deposit.ID, getDepositWithExternalRef.ID)
}
