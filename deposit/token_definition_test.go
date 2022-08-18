package deposit

import (
	"context"
	"encore.app/admin"
	"encore.app/commons/testutils"
	"encore.app/organization"
	"encore.dev/beta/errs"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateTokenDefinition(t *testing.T) {
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	orgSigningPubKey, _ := testutils.GenerateKeys()
	orgEncryptionPubKey, _ := testutils.GenerateKeys()
	notOrganizationPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:               testOrganizationId,
		Name:             testOrganizationId,
		SigningPubKey:    orgSigningPubKey,
		EncryptionPubKey: orgEncryptionPubKey,
	})
	require.NoError(t, err)

	testTable := []struct {
		name      string
		params    CreateTokenDefinitionParams
		errorCode errs.ErrCode
		uid       string
	}{
		{
			name: "Happy Path",
			params: CreateTokenDefinitionParams{
				OrganizationID: testOrganizationId,
				Name:           "My token def",
			},
			errorCode: errs.OK,
			uid:       orgSigningPubKey,
		},
		{
			name: "Organization does not exist",
			params: CreateTokenDefinitionParams{
				OrganizationID: "does not exist",
				Name:           "My token def",
			},
			errorCode: errs.NotFound,
			uid:       orgSigningPubKey,
		},
		{
			name: "Caller not organization",
			params: CreateTokenDefinitionParams{
				OrganizationID: testOrganizationId,
				Name:           "My voucher def",
			},
			errorCode: errs.PermissionDenied,
			uid:       notOrganizationPubKey,
		},
		{
			name: "Caller is admin",
			params: CreateTokenDefinitionParams{
				OrganizationID: testOrganizationId,
				Name:           "My token def",
			},
			errorCode: errs.OK,
			uid:       testutils.AdminPubKey,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, testutils.ClearDB(depositDB, "token_definition"))

			ctx := testutils.GetAuthenticatedContext(test.uid)
			resp, err := CreateTokenDefinition(ctx, &test.params)
			if test.errorCode == errs.OK {
				require.NoError(t, err)
				require.NotEqual(t, "", resp.ID)
				require.Equal(t, test.params.Name, resp.Name)
				require.Equal(t, test.params.OrganizationID, resp.OrganizationID)

				allTokenDefinitions, err := GetAllTokenDefinitions(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetAllTokenDefinitionsParams{})
				require.NoError(t, err)
				require.Equal(t, 1, len(allTokenDefinitions.TokenDefinitions))

				dbScheme, err := GetTokenDefinition(ctx, &GetTokenDefinitionParams{TokenDefinitionID: resp.ID})
				require.NoError(t, err)
				require.Equal(t, resp.ID, dbScheme.ID)
				require.Equal(t, test.params.Name, dbScheme.Name)
				require.Equal(t, test.params.OrganizationID, dbScheme.OrganizationID)
			} else {
				require.Error(t, err)
				require.Equal(t, test.errorCode, err.(*errs.Error).Code)

				allTokenDefinitions, err := GetAllTokenDefinitions(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetAllTokenDefinitionsParams{})
				require.NoError(t, err)
				require.Equal(t, 0, len(allTokenDefinitions.TokenDefinitions))
			}
		})
	}
}

func TestGetAllTokenDefinitions(t *testing.T) {
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	numberOfTokenDefs := 12
	tokenName := "MTN"

	require.NoError(t, admin.InsertTestData(context.Background()))
	orgSigningPubKey, _ := testutils.GenerateKeys()
	orgEncryptionPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:               testOrganizationId,
		Name:             testOrganizationId,
		SigningPubKey:    orgSigningPubKey,
		EncryptionPubKey: orgEncryptionPubKey,
	})

	for i := 0; i < numberOfTokenDefs; i++ {
		_, err := CreateTokenDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &CreateTokenDefinitionParams{
			Name:           tokenName,
			OrganizationID: testOrganizationId,
		})
		require.NoError(t, err)
	}

	allTokenDefinitions, err := GetAllTokenDefinitions(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetAllTokenDefinitionsParams{})
	require.NoError(t, err)
	require.Equal(t, numberOfTokenDefs, len(allTokenDefinitions.TokenDefinitions))

	for _, tokenDefinition := range allTokenDefinitions.TokenDefinitions {
		require.Equal(t, tokenName, tokenDefinition.Name)
		require.NotEqual(t, "", tokenDefinition.ID)
	}
}

func TestEditTokenDefinition(t *testing.T) {
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	tokenName := "MTN"

	require.NoError(t, admin.InsertTestData(context.Background()))
	orgSigningPubKey, _ := testutils.GenerateKeys()
	orgEncryptionPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:               testOrganizationId,
		Name:             testOrganizationId,
		SigningPubKey:    orgSigningPubKey,
		EncryptionPubKey: orgEncryptionPubKey,
	})

	vd, err := CreateTokenDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &CreateTokenDefinitionParams{
		Name:           tokenName,
		OrganizationID: testOrganizationId,
	})
	require.NoError(t, err)

	newName := "MY NewName is so cool"
	err = EditTokenDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &EditTokenDefinitionParams{
		TokenDefinitionID: vd.ID,
		Name:              newName,
	})
	require.NoError(t, err)

	vdAfterEdit, err := GetTokenDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetTokenDefinitionParams{
		TokenDefinitionID: vd.ID,
	})
	require.Equal(t, newName, vdAfterEdit.Name)
}
