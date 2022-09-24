package stats

import (
	"context"
	"testing"

	"encore.app/admin"
	"encore.app/commons"
	"encore.app/commons/testutils"
	"encore.app/deposit"
	"encore.app/organization"
	"encore.app/scheme"
	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/require"
)

const (
	testUserPubKey     = "testUserID"
	testOrganizationId = "myOrgId1"
)

var (
	depositDB                    = sqldb.Named("deposit")
	defaultTestRewardsMagnitude0 = commons.RewardDefinition{
		ItemDefinition: commons.ItemDefinition{
			MaterialDefinition: map[string]string{"materialType": "PET"},
			Magnitude:          commons.Weight,
		},
		RewardType:   commons.Voucher,
		RewardTypeID: "",
		PerItem:      2,
	}
	defaultTestRewardsMagnitude1 = commons.RewardDefinition{
		ItemDefinition: commons.ItemDefinition{
			MaterialDefinition: map[string]string{"materialType": "LDPE"},
			Magnitude:          commons.Count,
		},
		RewardType:   commons.Voucher,
		RewardTypeID: "",
		PerItem:      1,
	}
	defaultTestDeposit = []commons.MassBalance{
		{
			ItemDefinition: defaultTestRewardsMagnitude0.ItemDefinition,
			Amount:         12,
		},
	}
)

func TestGetStats(t *testing.T) {
	require.NoError(t, admin.InsertTestData(context.Background()))
	testutils.ClearAllDBs()

	user1, _ := testutils.GenerateKeys()

	context := testutils.GetAuthenticatedContext(user1)

	user1Stats, err := GetStats(context, &User{
		PubKey: user1,
	})

	// User empty data
	require.Equal(t, int64(0), user1Stats.NumberOfAvailableVouchers)
	require.Equal(t, float64(0), user1Stats.PlasticCollected)
	require.Equal(t, int64(0), user1Stats.NumberOfUsedVouchers)
	require.Equal(t, 0, len(user1Stats.DepositAmounts))

	// pubkeys
	orgSigningPubKey, _ := testutils.GenerateKeys()
	orgEncryptionPubKey, _ := testutils.GenerateKeys()

	// Organization added
	_, err = organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:               testOrganizationId,
		Name:             testOrganizationId,
		SigningPubKey:    orgSigningPubKey,
		EncryptionPubKey: orgEncryptionPubKey,
	})
	require.NoError(t, err)

	// Defines voucher
	definition, err := deposit.CreateVoucherDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &deposit.CreateVoucherDefinitionParams{
		OrganizationID: testOrganizationId,
		Name:           "Voucher def name",
		PictureURL:     "https://does.not.matter.com",
	})
	require.NoError(t, err)

	defaultTestRewardsMagnitude1.RewardTypeID = definition.ID

	collectionPointPubKey, _ := testutils.GenerateKeys()
	// Adds scheme
	testScheme, err := scheme.CreateScheme(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &scheme.CreateSchemeParams{
		Name: "TestScheme",
		RewardDefinitions: []commons.RewardDefinition{
			defaultTestRewardsMagnitude1,
		},
		OrganizationID: testOrganizationId,
	})
	require.NoError(t, err)

	// Add Collection point
	err = scheme.AddCollectionPoint(testutils.GetAuthenticatedContext(orgSigningPubKey), &scheme.AddCollectionPointParams{
		SchemeID:              testScheme.ID,
		CollectionPointPubKey: collectionPointPubKey,
	})
	require.NoError(t, err)

	// Just Context
	contextDeposit := testutils.GetAuthenticatedContext(collectionPointPubKey)

	// First Deposit made
	deposit1, err := deposit.MakeDeposit(contextDeposit, &deposit.MakeDepositParams{
		SchemeID:   testScheme.ID,
		UserPubKey: user1,
		MassBalanceDeposits: []commons.MassBalance{
			{
				ItemDefinition: defaultTestRewardsMagnitude1.ItemDefinition,
				Amount:         20,
			},
		},
	})
	require.NoError(t, err)

	userStatsDeposit1, err := GetStats(context, &User{
		PubKey: deposit1.UserPubKey,
	})

	// User First deposit
	require.Equal(t, int64(20), userStatsDeposit1.NumberOfAvailableVouchers) //  1 item = 1 voucher
	require.Equal(t, float64(0), userStatsDeposit1.PlasticCollected)
	require.Equal(t, 1, len(userStatsDeposit1.DepositAmounts))
	require.Equal(t, int64(0), userStatsDeposit1.NumberOfUsedVouchers) // this doesn't change

	// Defines second voucher
	definition, err = deposit.CreateVoucherDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &deposit.CreateVoucherDefinitionParams{
		OrganizationID: testOrganizationId,
		Name:           "Voucher 2",
		PictureURL:     "https://does.not.matter.com",
	})
	require.NoError(t, err)

	defaultTestRewardsMagnitude0.RewardTypeID = definition.ID

	// Adds scheme
	testScheme, err = scheme.CreateScheme(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &scheme.CreateSchemeParams{
		Name: "TestScheme 2",
		RewardDefinitions: []commons.RewardDefinition{
			defaultTestRewardsMagnitude0,
		},
		OrganizationID: testOrganizationId,
	})
	require.NoError(t, err)

	// Add Collection point
	err = scheme.AddCollectionPoint(testutils.GetAuthenticatedContext(orgSigningPubKey), &scheme.AddCollectionPointParams{
		SchemeID:              testScheme.ID,
		CollectionPointPubKey: collectionPointPubKey,
	})
	require.NoError(t, err)

	// Just Context

	// First Deposit made
	_, err = deposit.MakeDeposit(contextDeposit, &deposit.MakeDepositParams{
		SchemeID:   testScheme.ID,
		UserPubKey: user1,
		MassBalanceDeposits: []commons.MassBalance{
			{
				ItemDefinition: defaultTestRewardsMagnitude0.ItemDefinition,
				Amount:         5,
			},
		},
	})
	require.NoError(t, err)

	userStatsDeposit2, err := GetStats(context, &User{
		PubKey: deposit1.UserPubKey,
	})

	// User First deposit
	require.Equal(t, int64(30), userStatsDeposit2.NumberOfAvailableVouchers) // 20 + 5*2 (1 item = 2 vouchers)
	require.Equal(t, float64(5), userStatsDeposit2.PlasticCollected)         // 5 kg
	require.Equal(t, 2, len(userStatsDeposit2.DepositAmounts))
	require.Equal(t, int64(0), userStatsDeposit2.NumberOfUsedVouchers) // this doesn't change

	allVouchers, _ := deposit.GetVouchersForUser(testutils.GetAuthenticatedContext(user1), &deposit.GetVouchersForUserParams{UserPubKey: user1})

	var numberOfVouchersAvailable = int64(30)
	var numberOfVouchersUsed = int64(0)
	for _, voucher := range allVouchers.Vouchers {
		voucherId := voucher.Voucher.ID
		deposit.InvalidateVoucher(context, &deposit.InvalidateVoucherParams{VoucherID: voucherId})
		userStatsUsingVouchers, _ := GetStats(context, &User{
			PubKey: user1,
		})
		numberOfVouchersAvailable -= 1
		numberOfVouchersUsed += 1

		// User First deposit
		require.Equal(t, numberOfVouchersAvailable, userStatsUsingVouchers.NumberOfAvailableVouchers)
		require.Equal(t, float64(5), userStatsUsingVouchers.PlasticCollected)
		require.Equal(t, 2, len(userStatsUsingVouchers.DepositAmounts))
		require.Equal(t, numberOfVouchersUsed, userStatsUsingVouchers.NumberOfUsedVouchers) // this doesn't change
	}

}
