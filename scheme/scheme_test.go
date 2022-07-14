package scheme

import (
	"context"
	"encore.app/admin"
	"encore.app/commons"
	"encore.app/commons/testutils"
	"encore.app/organization"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/require"
	"testing"
)

const testOrganizationId = "myOrgId1"

var (
	schemeDB           = sqldb.Named("scheme")
	defaultTestRewards = []commons.RewardDefinition{
		{
			ItemDefinition: commons.ItemDefinition{
				MaterialDefinition: map[string]string{"materialType": "PET"},
				Magnitude:          commons.Weight,
			},
			RewardType:    commons.Token,
			RewardPerUnit: 1,
		},
	}
)

func TestCreateScheme(t *testing.T) {
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	organizationPubKey, _ := testutils.GenerateKeys()
	notOrganizationPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:     testOrganizationId,
		Name:   testOrganizationId,
		PubKey: organizationPubKey,
	})
	require.NoError(t, err)

	testTable := []struct {
		name      string
		params    CreateSchemeParams
		errorCode errs.ErrCode
		uid       string
	}{
		{
			name: "Happy path",
			params: CreateSchemeParams{
				Name:              "Valid",
				OrganizationID:    testOrganizationId,
				RewardDefinitions: defaultTestRewards,
			},
			errorCode: errs.OK,
			uid:       organizationPubKey,
		},
		{
			name: "Invalid: empty name",
			params: CreateSchemeParams{
				Name:              "",
				OrganizationID:    testOrganizationId,
				RewardDefinitions: defaultTestRewards,
			},
			errorCode: errs.InvalidArgument,
			uid:       organizationPubKey,
		},
		// TODO: TEST MORE PARAMS
		{
			name: "Unauthenticated",
			params: CreateSchemeParams{
				Name:              "Valid",
				OrganizationID:    testOrganizationId,
				RewardDefinitions: defaultTestRewards,
			},
			errorCode: errs.Unauthenticated,
			uid:       "",
		},
		{
			name: "Organization doesnt exist",
			params: CreateSchemeParams{
				Name:              "Valid",
				OrganizationID:    "does not exist",
				RewardDefinitions: defaultTestRewards,
			},
			errorCode: errs.NotFound,
			uid:       organizationPubKey,
		},
		{
			name: "Not done by organization",
			params: CreateSchemeParams{
				Name:              "Valid",
				OrganizationID:    testOrganizationId,
				RewardDefinitions: defaultTestRewards,
			},
			errorCode: errs.PermissionDenied,
			uid:       notOrganizationPubKey,
		},
		{
			name: "By admin",
			params: CreateSchemeParams{
				Name:              "Valid",
				OrganizationID:    testOrganizationId,
				RewardDefinitions: defaultTestRewards,
			},
			errorCode: errs.OK,
			uid:       testutils.AdminPubKey,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			var ctx context.Context
			if test.uid != "" {
				ctx = testutils.GetAuthenticatedContext(test.uid)
			} else {
				ctx = context.Background()
			}
			resp, err := CreateScheme(ctx, &test.params)
			if test.errorCode == errs.OK {
				require.NoError(t, err)
				require.NotEqual(t, "", resp.ID)

				getAllResp, err := GetAllSchemes(testutils.GetAuthenticatedContext(""))
				require.NoError(t, err)
				require.Equal(t, 1, len(getAllResp.Schemes))

				dbScheme, err := GetScheme(ctx, &GetSchemeParams{SchemeID: resp.ID})
				require.NoError(t, err)
				require.Equal(t, resp.ID, dbScheme.ID)
				require.Equal(t, test.params.Name, dbScheme.Name)
				require.Equal(t, test.params.OrganizationID, dbScheme.OrganizationID)
				require.Equal(t, len(test.params.RewardDefinitions), len(dbScheme.RewardDefinitions))
				require.True(t, test.params.RewardDefinitions[0].ItemDefinition.SameAs(dbScheme.RewardDefinitions[0].ItemDefinition))
			} else {
				require.Error(t, err)
				require.Equal(t, test.errorCode, err.(*errs.Error).Code)

				getAllResp, err := GetAllSchemes(testutils.GetAuthenticatedContext(""))
				require.NoError(t, err)
				require.Equal(t, 0, len(getAllResp.Schemes))
			}

			require.NoError(t, testutils.ClearDB(schemeDB, "scheme"))
		})
	}
}
