package deposit

import (
	"context"
	"encore.app/admin"
	"encore.app/commons"
	"encore.app/commons/testutils"
	"encore.app/organization"
	"encore.app/scheme"
	"encore.dev/beta/errs"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClaim(t *testing.T) {
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	userPubKey, _ := testutils.GenerateKeys()
	notUserPubKey, _ := testutils.GenerateKeys()

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

	testTable := []struct {
		name           string
		userPubKey     string
		errorCode      errs.ErrCode
		useRealDeposit bool
		uid            string
	}{
		{
			name:           "Happy path: user claim",
			userPubKey:     userPubKey,
			errorCode:      errs.OK,
			useRealDeposit: true,
			uid:            userPubKey,
		},
		{
			name:           "Happy path: collection point claim for user",
			userPubKey:     userPubKey,
			errorCode:      errs.OK,
			useRealDeposit: true,
			uid:            collectionPointPubKey,
		},
		{
			name:           "Happy path: admin claim for user",
			userPubKey:     userPubKey,
			errorCode:      errs.OK,
			useRealDeposit: true,
			uid:            testutils.AdminPubKey,
		},
		{
			name:           "Claim for another user",
			userPubKey:     userPubKey,
			errorCode:      errs.PermissionDenied,
			useRealDeposit: true,
			uid:            notUserPubKey,
		},
		{
			name:           "Deposit not found",
			userPubKey:     userPubKey,
			errorCode:      errs.NotFound,
			useRealDeposit: false,
			uid:            userPubKey,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			deposit, err := MakeDeposit(testutils.GetAuthenticatedContext(collectionPointPubKey), &MakeDepositParams{
				SchemeID:            testScheme.ID,
				MassBalanceDeposits: defaultTestDeposit,
			})
			require.NoError(t, err)

			ctx := testutils.GetAuthenticatedContext(test.uid)

			depositID := deposit.ID
			if !test.useRealDeposit {
				depositID = "does not exist"
			}

			resp, err := Claim(ctx, &ClaimParams{
				DepositID:  depositID,
				UserPubKey: test.userPubKey,
			})
			_ = resp
			if test.errorCode == errs.OK {
				require.NoError(t, err)

				dbDeposit, err := GetDeposit(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetDepositParams{DepositID: deposit.ID})
				require.NoError(t, err)
				require.True(t, dbDeposit.Claimed)
				require.Equal(t, userPubKey, dbDeposit.UserPubKey)

				getVouchersResp, err := GetVouchersForUser(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetVouchersForUserParams{UserPubKey: test.userPubKey})
				require.NoError(t, err)
				require.Equal(t, 12, len(getVouchersResp.Vouchers))
			} else {
				require.Error(t, err)
				require.Equal(t, test.errorCode, err.(*errs.Error).Code)

				dbDeposit, err := GetDeposit(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetDepositParams{DepositID: deposit.ID})
				require.NoError(t, err)
				require.False(t, dbDeposit.Claimed)

				getVouchersResp, err := GetVouchersForUser(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetVouchersForUserParams{UserPubKey: test.userPubKey})
				require.NoError(t, err)
				require.Equal(t, 0, len(getVouchersResp.Vouchers))
			}

			require.NoError(t, testutils.ClearDB(depositDB, "deposit", "voucher"))
		})
	}
}

func TestDoubleClaim(t *testing.T) {
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	userPubKey, _ := testutils.GenerateKeys()

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

	deposit, err := MakeDeposit(testutils.GetAuthenticatedContext(collectionPointPubKey), &MakeDepositParams{
		SchemeID:            testScheme.ID,
		MassBalanceDeposits: defaultTestDeposit,
	})
	require.NoError(t, err)

	_, err = Claim(testutils.GetAuthenticatedContext(userPubKey), &ClaimParams{
		DepositID:  deposit.ID,
		UserPubKey: userPubKey,
	})
	require.NoError(t, err)

	_, err = Claim(testutils.GetAuthenticatedContext(userPubKey), &ClaimParams{
		DepositID:  deposit.ID,
		UserPubKey: userPubKey,
	})
	require.Error(t, err)
	require.Equal(t, errs.InvalidArgument, err.(*errs.Error).Code)
}
