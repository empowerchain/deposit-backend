package commons

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

const defaultRewardTypeID = "whatever"

var defaultItemDefinition = ItemDefinition{
	MaterialDefinition: map[string]string{"materialType": "PET"},
	Magnitude:          Weight,
}

func TestGetRewardsFor(t *testing.T) {
	testTable := []struct {
		rewardDef RewardDefinition
		deposit   MassBalance
		expected  Reward
	}{
		{
			rewardDef: RewardDefinition{
				ItemDefinition: defaultItemDefinition,
				RewardType:     Token,
				RewardTypeID:   defaultRewardTypeID,
				PerItem:        1,
			},
			deposit: MassBalance{
				ItemDefinition: defaultItemDefinition,
				Amount:         5,
			},
			expected: Reward{
				Type:   Token,
				TypeID: defaultRewardTypeID,
				Amount: 5,
			},
		},
		{
			rewardDef: RewardDefinition{
				ItemDefinition: defaultItemDefinition,
				RewardType:     Voucher,
				RewardTypeID:   defaultRewardTypeID,
				PerItem:        0.1,
			},
			deposit: MassBalance{
				ItemDefinition: defaultItemDefinition,
				Amount:         21,
			},
			expected: Reward{
				Type:   Voucher,
				TypeID: defaultRewardTypeID,
				Amount: 2.1,
			},
		},
		{
			rewardDef: RewardDefinition{
				ItemDefinition: defaultItemDefinition,
				RewardType:     Voucher,
				RewardTypeID:   defaultRewardTypeID,
				PerItem:        1.99999999,
			},
			deposit: MassBalance{
				ItemDefinition: defaultItemDefinition,
				Amount:         0.33333339,
			},
			expected: Reward{
				Type:   Voucher,
				TypeID: defaultRewardTypeID,
				Amount: 0.6666667766666661,
			},
		},
	}

	for _, test := range testTable {
		testName := fmt.Sprintf("Given PerItem=%f, When Deposit.Amount=%f, Then Reward.Amount should be %f", test.rewardDef.PerItem, test.deposit.Amount, test.expected.Amount)
		t.Run(testName, func(t *testing.T) {
			actual := test.rewardDef.GetRewardsFor(test.deposit)
			require.Equal(t, test.expected.Type, actual.Type)
			require.Equal(t, test.expected.TypeID, actual.TypeID)

			diff := math.Abs(test.expected.Amount - actual.Amount)
			require.True(t, diff < 0.000000000001) // Very small difference
		})
	}
}
