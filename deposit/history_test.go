package deposit

import (
	"context"
	"encore.app/admin"
	"encore.app/commons"
	"encore.app/commons/testutils"
	"encore.app/organization"
	"encore.app/scheme"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHistory(t *testing.T) {
	require.NoError(t, admin.InsertTestData(context.Background()))
	testutils.ClearAllDBs()

	user1, _ := testutils.GenerateKeys()
	user2, _ := testutils.GenerateKeys()
	user3, _ := testutils.GenerateKeys()

	orgSigningPubKey, _ := testutils.GenerateKeys()
	orgEncryptionPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:               testOrganizationId,
		Name:             testOrganizationId,
		SigningPubKey:    orgSigningPubKey,
		EncryptionPubKey: orgEncryptionPubKey,
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

	err = scheme.AddCollectionPoint(testutils.GetAuthenticatedContext(orgSigningPubKey), &scheme.AddCollectionPointParams{
		SchemeID:              testScheme.ID,
		CollectionPointPubKey: collectionPointPubKey,
	})
	require.NoError(t, err)

	ctx := testutils.GetAuthenticatedContext(collectionPointPubKey)
	user1Deposit1, err := MakeDeposit(ctx, &MakeDepositParams{
		SchemeID:   testScheme.ID,
		UserPubKey: user1,
		MassBalanceDeposits: []commons.MassBalance{
			{
				ItemDefinition: defaultTestRewards.ItemDefinition,
				Amount:         12,
			},
		},
	})
	require.NoError(t, err)

	user1Deposit2, err := MakeDeposit(ctx, &MakeDepositParams{
		SchemeID:   testScheme.ID,
		UserPubKey: user1,
		MassBalanceDeposits: []commons.MassBalance{
			{
				ItemDefinition: defaultTestRewards.ItemDefinition,
				Amount:         7,
			},
		},
	})
	require.NoError(t, err)

	user2Deposit1, err := MakeDeposit(ctx, &MakeDepositParams{
		SchemeID:   testScheme.ID,
		UserPubKey: user2,
		MassBalanceDeposits: []commons.MassBalance{
			{
				ItemDefinition: defaultTestRewards.ItemDefinition,
				Amount:         42,
			},
		},
	})
	require.NoError(t, err)

	user1History, err := GetHistory(ctx, &GetHistoryParams{
		UserPubKey: user1,
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(user1History.Events))
	// user1Deposit2 first because newest first
	require.Equal(t, user1Deposit2.MassBalanceDeposits[0].Amount, user1History.Events[0].NumberOfUnitsIn)
	require.Equal(t, EventTypeDeposit, user1History.Events[0].EventType)
	require.Equal(t, "Voucher", user1History.Events[0].UnitNameIn)

	require.Equal(t, user1Deposit1.MassBalanceDeposits[0].Amount, user1History.Events[1].NumberOfUnitsIn)
	require.Equal(t, EventTypeDeposit, user1History.Events[1].EventType)
	require.Equal(t, "Voucher", user1History.Events[1].UnitNameIn)

	user2History, err := GetHistory(ctx, &GetHistoryParams{
		UserPubKey: user2,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(user2History.Events))
	require.Equal(t, user2Deposit1.MassBalanceDeposits[0].Amount, user2History.Events[0].NumberOfUnitsIn)
	require.Equal(t, EventTypeDeposit, user2History.Events[0].EventType)
	require.Equal(t, "Voucher", user2History.Events[0].UnitNameIn)

	user3History, err := GetHistory(ctx, &GetHistoryParams{
		UserPubKey: user3,
	})
	require.NoError(t, err)
	require.Equal(t, 0, len(user3History.Events))
}
