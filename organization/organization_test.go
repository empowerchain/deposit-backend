package organization

import (
	"context"
	"encore.app/admin"
	"encore.app/commons/testutils"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

var organizationDB = sqldb.Named("organization")

func TestCreateOrganization(t *testing.T) {
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))
	pubKey, _ := testutils.GenerateKeys()
	notAdmin, _ := testutils.GenerateKeys()

	testTable := []struct {
		name      string
		uid       string
		params    CreateOrgParams
		errorCode errs.ErrCode
	}{
		{
			name: "Happy pattestutils.ClearAllDBs()h",
			uid:  testutils.AdminPubKey,
			params: CreateOrgParams{
				ID:     "testId1",
				Name:   "My Org",
				PubKey: pubKey,
			},
			errorCode: errs.OK,
		},
		{
			name: "Not admin",
			uid:  notAdmin,
			params: CreateOrgParams{
				ID:     "testId1",
				Name:   "My Org",
				PubKey: pubKey,
			},
			errorCode: errs.PermissionDenied,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			resp, err := CreateOrganization(testutils.GetAuthenticatedContext(test.uid), &test.params)
			if test.errorCode == errs.OK {
				require.NoError(t, err)

				require.Equal(t, test.params.ID, resp.ID)
				require.Equal(t, test.params.Name, resp.Name)
				require.Equal(t, test.params.PubKey, resp.PubKey)

				allOrgs, err := GetAllOrganizations(testutils.GetAuthenticatedContext(testutils.AdminPubKey))
				require.NoError(t, err)
				require.Equal(t, 1, len(allOrgs.Organizations))

				orgDb, err := GetOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetOrganizationParams{ID: test.params.ID})
				require.NoError(t, err)
				require.Equal(t, test.params.ID, orgDb.ID)
				require.Equal(t, test.params.Name, orgDb.Name)
				require.Equal(t, test.params.PubKey, orgDb.PubKey)
			} else {
				require.Error(t, err)
				require.Equal(t, test.errorCode, err.(*errs.Error).Code)

				allOrgs, err := GetAllOrganizations(testutils.GetAuthenticatedContext(testutils.AdminPubKey))
				require.NoError(t, err)
				require.Equal(t, 0, len(allOrgs.Organizations))
			}

			require.NoError(t, testutils.ClearDB(organizationDB, "organization"))
		})
	}
}

func TestGetAllOrganizations(t *testing.T) {
	numberOfOrgs := 21
	orgName := "My name"

	require.NoError(t, admin.InsertTestData(context.Background()))
	testutils.ClearAllDBs()

	for i := 0; i < numberOfOrgs; i++ {
		pubKey, _ := testutils.GenerateKeys()

		_, err := CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &CreateOrgParams{
			ID:     strconv.Itoa(i),
			Name:   orgName,
			PubKey: pubKey,
		})
		require.NoError(t, err)
	}

	allOrgs, err := GetAllOrganizations(testutils.GetAuthenticatedContext(testutils.AdminPubKey))
	require.NoError(t, err)
	require.Equal(t, numberOfOrgs, len(allOrgs.Organizations))

	for i, org := range allOrgs.Organizations {
		require.Equal(t, strconv.Itoa(i), org.ID)
		require.Equal(t, orgName, org.Name)
		require.NotEqual(t, "", org.PubKey)
	}
}
