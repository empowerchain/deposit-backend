package deposit

import (
	"context"
	"testing"

	"encore.app/admin"
	"encore.app/commons/testutils"
	"encore.app/organization"
	"encore.dev/beta/errs"
	"github.com/stretchr/testify/require"
)

func TestCreateVoucherDefinition(t *testing.T) {
	testutils.EnsureExclusiveDatabaseAccess(t)
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
		params    CreateVoucherDefinitionParams
		errorCode errs.ErrCode
		uid       string
	}{
		{
			name: "Happy Path",
			params: CreateVoucherDefinitionParams{
				OrganizationID: testOrganizationId,
				Name:           "My voucher def",
				PictureURL:     "https://whatever.com/image.png",
			},
			errorCode: errs.OK,
			uid:       orgSigningPubKey,
		},
		{
			name: "Organization does not exist",
			params: CreateVoucherDefinitionParams{
				OrganizationID: "does not exist",
				Name:           "My voucher def",
				PictureURL:     "https://whatever.com/image.png",
			},
			errorCode: errs.NotFound,
			uid:       orgSigningPubKey,
		},
		{
			name: "Caller not organization",
			params: CreateVoucherDefinitionParams{
				OrganizationID: testOrganizationId,
				Name:           "My voucher def",
				PictureURL:     "https://whatever.com/image.png",
			},
			errorCode: errs.PermissionDenied,
			uid:       notOrganizationPubKey,
		},
		{
			name: "Caller is admin",
			params: CreateVoucherDefinitionParams{
				OrganizationID: testOrganizationId,
				Name:           "My voucher def",
				PictureURL:     "https://whatever.com/image.png",
			},
			errorCode: errs.OK,
			uid:       testutils.AdminPubKey,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, testutils.ClearDB(depositDB, "voucher_definition"))

			ctx := testutils.GetAuthenticatedContext(test.uid)
			resp, err := CreateVoucherDefinition(ctx, &test.params)
			if test.errorCode == errs.OK {
				require.NoError(t, err)
				require.NotEqual(t, "", resp.ID)
				require.Equal(t, test.params.Name, resp.Name)
				require.Equal(t, test.params.OrganizationID, resp.OrganizationID)
				require.Equal(t, test.params.PictureURL, resp.PictureURL)

				allVoucherDefinitions, err := GetAllVoucherDefinitions(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetAllVoucherDefinitionsParams{})
				require.NoError(t, err)
				require.Equal(t, 1, len(allVoucherDefinitions.VoucherDefinitions))

				dbScheme, err := GetVoucherDefinition(ctx, &GetVoucherDefinitionParams{VoucherDefinitionID: resp.ID})
				require.NoError(t, err)
				require.Equal(t, resp.ID, dbScheme.ID)
				require.Equal(t, test.params.Name, dbScheme.Name)
				require.Equal(t, test.params.OrganizationID, dbScheme.OrganizationID)
				require.Equal(t, test.params.PictureURL, dbScheme.PictureURL)
			} else {
				require.Error(t, err)
				require.Equal(t, test.errorCode, err.(*errs.Error).Code)

				allVoucherDefinitions, err := GetAllVoucherDefinitions(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetAllVoucherDefinitionsParams{})
				require.NoError(t, err)
				require.Equal(t, 0, len(allVoucherDefinitions.VoucherDefinitions))
			}
		})
	}
}

func TestGetAllVoucherDefinitions(t *testing.T) {
	testutils.EnsureExclusiveDatabaseAccess(t)
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	numberOfVoucherDefs := 12
	voucherName := "My voucher name"
	pictureURL := "https://something.co/pop.png"

	require.NoError(t, admin.InsertTestData(context.Background()))
	orgSigningPubKey, _ := testutils.GenerateKeys()
	orgEncryptionPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:               testOrganizationId,
		Name:             testOrganizationId,
		SigningPubKey:    orgSigningPubKey,
		EncryptionPubKey: orgEncryptionPubKey,
	})

	for i := 0; i < numberOfVoucherDefs; i++ {
		_, err := CreateVoucherDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &CreateVoucherDefinitionParams{
			Name:           voucherName,
			OrganizationID: testOrganizationId,
			PictureURL:     pictureURL,
		})
		require.NoError(t, err)
	}

	allVoucherDefinitions, err := GetAllVoucherDefinitions(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetAllVoucherDefinitionsParams{})
	require.NoError(t, err)
	require.Equal(t, numberOfVoucherDefs, len(allVoucherDefinitions.VoucherDefinitions))

	for _, voucherDefinition := range allVoucherDefinitions.VoucherDefinitions {
		require.Equal(t, voucherName, voucherDefinition.Name)
		require.Equal(t, pictureURL, voucherDefinition.PictureURL)
		require.NotEqual(t, "", voucherDefinition.ID)
	}
}

func TestEditVoucherDefinition(t *testing.T) {
	testutils.EnsureExclusiveDatabaseAccess(t)
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	voucherName := "My voucher name"
	pictureURL := "https://something.co/pop.png"

	require.NoError(t, admin.InsertTestData(context.Background()))
	orgSigningPubKey, _ := testutils.GenerateKeys()
	orgEncryptionPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:               testOrganizationId,
		Name:             testOrganizationId,
		SigningPubKey:    orgSigningPubKey,
		EncryptionPubKey: orgEncryptionPubKey,
	})

	vd, err := CreateVoucherDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &CreateVoucherDefinitionParams{
		Name:           voucherName,
		OrganizationID: testOrganizationId,
		PictureURL:     pictureURL,
	})
	require.NoError(t, err)

	newName := "MY NewName is so cool"
	newPictureURL := "https://mypicture.horse.com/coolhorse.jpeg"
	err = EditVoucherDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &EditVoucherDefinitionParams{
		VoucherDefinitionID: vd.ID,
		Name:                newName,
		PictureURL:          newPictureURL,
	})
	require.NoError(t, err)

	vdAfterEdit, err := GetVoucherDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetVoucherDefinitionParams{
		VoucherDefinitionID: vd.ID,
	})
	require.Equal(t, newName, vdAfterEdit.Name)
	require.Equal(t, newPictureURL, vdAfterEdit.PictureURL)
}
