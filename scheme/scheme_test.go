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
			RewardType:   commons.Token,
			RewardTypeID: "whatever",
			PerItem:      1,
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
			require.NoError(t, testutils.ClearDB(schemeDB, "scheme"))

			ctx := testutils.GetAuthenticatedContext(test.uid)
			resp, err := CreateScheme(ctx, &test.params)
			if test.errorCode == errs.OK {
				require.NoError(t, err)
				require.NotEqual(t, "", resp.ID)

				getAllResp, err := GetAllSchemes(testutils.GetAuthenticatedContext(""), &GetAllSchemesParams{})
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

				getAllResp, err := GetAllSchemes(testutils.GetAuthenticatedContext(""), &GetAllSchemesParams{})
				require.NoError(t, err)
				require.Equal(t, 0, len(getAllResp.Schemes))
			}
		})
	}
}

func TestAddCollectionPoint(t *testing.T) {
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
		name        string
		useSchemeID bool
		errorCode   errs.ErrCode
		uid         string
	}{
		{
			name:        "Happy Path",
			useSchemeID: true,
			errorCode:   errs.OK,
			uid:         organizationPubKey,
		},
		{
			name:        "Unauthorized",
			useSchemeID: true,
			errorCode:   errs.PermissionDenied,
			uid:         notOrganizationPubKey,
		},
		{
			name:        "Admin",
			useSchemeID: true,
			errorCode:   errs.OK,
			uid:         testutils.AdminPubKey,
		},
		{
			name:        "Not found",
			useSchemeID: false,
			errorCode:   errs.NotFound,
			uid:         organizationPubKey,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, testutils.ClearDB(schemeDB, "scheme"))

			scheme, err := CreateScheme(testutils.GetAuthenticatedContext(organizationPubKey), &CreateSchemeParams{
				Name:              "SchemeName",
				OrganizationID:    testOrganizationId,
				RewardDefinitions: defaultTestRewards,
			})
			require.NoError(t, err)

			ctx := testutils.GetAuthenticatedContext(test.uid)

			schemeID := scheme.ID
			if !test.useSchemeID {
				schemeID = "something else"
			}
			collectionPointsPubKey, _ := testutils.GenerateKeys()
			err = AddCollectionPoint(ctx, &AddCollectionPointParams{
				SchemeID:              schemeID,
				CollectionPointPubKey: collectionPointsPubKey,
			})
			if test.errorCode == errs.OK {
				require.NoError(t, err)

				dbScheme, err := GetScheme(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetSchemeParams{SchemeID: scheme.ID})
				require.NoError(t, err)
				require.Equal(t, 1, len(dbScheme.CollectionPoints))
				require.Equal(t, collectionPointsPubKey, dbScheme.CollectionPoints[0])
			} else {
				require.Error(t, err)
				require.Equal(t, test.errorCode, err.(*errs.Error).Code)

				dbScheme, err := GetScheme(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetSchemeParams{SchemeID: scheme.ID})
				require.NoError(t, err)
				require.Equal(t, 0, len(dbScheme.CollectionPoints))
			}
		})
	}
}

func TestEditScheme(t *testing.T) {
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	organizationPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:     testOrganizationId,
		Name:   testOrganizationId,
		PubKey: organizationPubKey,
	})
	require.NoError(t, err)

	scheme, err := CreateScheme(testutils.GetAuthenticatedContext(organizationPubKey), &CreateSchemeParams{
		Name:              "SchemeName",
		OrganizationID:    testOrganizationId,
		RewardDefinitions: defaultTestRewards,
	})
	require.NoError(t, err)

	newRewardDef := commons.RewardDefinition{
		ItemDefinition: commons.ItemDefinition{
			MaterialDefinition: map[string]string{"plasticType": "LDPE"},
			Magnitude:          commons.Weight,
		},
		RewardType:   commons.Voucher,
		RewardTypeID: "whatever2",
		PerItem:      1.5,
	}

	collectionPoint1, _ := testutils.GenerateKeys()
	collectionPoint2, _ := testutils.GenerateKeys()
	err = EditScheme(testutils.GetAuthenticatedContext(organizationPubKey), &EditSchemeParams{
		SchemeID: scheme.ID,
		RewardDefinitions: []commons.RewardDefinition{
			newRewardDef,
		},
		CollectionPoints: []string{
			collectionPoint1,
			collectionPoint2,
		},
	})
	require.NoError(t, err)

	dbScheme, err := GetScheme(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetSchemeParams{SchemeID: scheme.ID})
	require.NoError(t, err)
	require.Equal(t, 1, len(dbScheme.RewardDefinitions))
	require.True(t, newRewardDef.ItemDefinition.SameAs(dbScheme.RewardDefinitions[0].ItemDefinition))
	require.Equal(t, newRewardDef.RewardType, dbScheme.RewardDefinitions[0].RewardType)
	require.Equal(t, newRewardDef.RewardTypeID, dbScheme.RewardDefinitions[0].RewardTypeID)
	require.Equal(t, newRewardDef.PerItem, dbScheme.RewardDefinitions[0].PerItem)
	require.Equal(t, 2, len(dbScheme.CollectionPoints))
	require.Equal(t, collectionPoint1, dbScheme.CollectionPoints[0])
	require.Equal(t, collectionPoint2, dbScheme.CollectionPoints[1])
}
