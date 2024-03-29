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

func TestGetAllVouchers(t *testing.T) {
	testutils.EnsureExclusiveDatabaseAccess(t)
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	numberOfVouchers := 3

	orgSigningPubKey, _ := testutils.GenerateKeys()
	orgEncryptionPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:               testOrganizationId,
		Name:             testOrganizationId,
		SigningPubKey:    orgSigningPubKey,
		EncryptionPubKey: orgEncryptionPubKey,
	})
	require.NoError(t, err)

	voucherDefinition, err := CreateVoucherDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &CreateVoucherDefinitionParams{
		OrganizationID: testOrganizationId,
		Name:           "My voucher",
		PictureURL:     "https://whatever.com/pic.jpeg",
	})

	tx, err := depositDB.Begin(context.Background())
	require.NoError(t, err)
	for i := 0; i < numberOfVouchers; i++ {
		pubKey, _ := testutils.GenerateKeys()

		_, err := mintVoucher(testutils.GetAuthenticatedContext(testutils.AdminPubKey), tx, voucherDefinition, pubKey)
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())

	allVouchers, err := GetAllVouchers(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetAllVouchersParams{})
	require.NoError(t, err)
	require.Equal(t, numberOfVouchers, len(allVouchers.Vouchers))

	for _, voucherRes := range allVouchers.Vouchers {
		require.Equal(t, voucherDefinition.ID, voucherRes.Voucher.VoucherDefinitionID)
		require.NotEqual(t, "", voucherRes.Voucher.ID)
		require.NotEqual(t, "pictureURL", voucherRes.Voucher.OwnerPubKey)
	}
}

func TestGetAllVouchersForUser(t *testing.T) {
	testutils.EnsureExclusiveDatabaseAccess(t)
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	userPubKey, _ := testutils.GenerateKeys()
	numberOfVouchersForUser := 3
	otherUserPubKey, _ := testutils.GenerateKeys()
	numberOfVouchersForOtherUser := 7

	orgSigningPubKey, _ := testutils.GenerateKeys()
	orgEncryptionPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:               testOrganizationId,
		Name:             testOrganizationId,
		SigningPubKey:    orgSigningPubKey,
		EncryptionPubKey: orgEncryptionPubKey,
	})
	require.NoError(t, err)

	voucherDefinition, err := CreateVoucherDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &CreateVoucherDefinitionParams{
		OrganizationID: testOrganizationId,
		Name:           "My voucher",
		PictureURL:     "https://whatever.com/pic.jpeg",
	})

	tx, err := depositDB.Begin(context.Background())
	require.NoError(t, err)
	for i := 0; i < numberOfVouchersForUser; i++ {
		_, err := mintVoucher(testutils.GetAuthenticatedContext(testutils.AdminPubKey), tx, voucherDefinition, userPubKey)
		require.NoError(t, err)
	}

	for i := 0; i < numberOfVouchersForOtherUser; i++ {
		_, err := mintVoucher(testutils.GetAuthenticatedContext(testutils.AdminPubKey), tx, voucherDefinition, otherUserPubKey)
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit())

	allVouchers, err := GetAllVouchers(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetAllVouchersParams{})
	require.NoError(t, err)
	require.Equal(t, numberOfVouchersForUser+numberOfVouchersForOtherUser, len(allVouchers.Vouchers))

	vouchersForUser, err := GetVouchersForUser(testutils.GetAuthenticatedContext(userPubKey), &GetVouchersForUserParams{UserPubKey: userPubKey})
	require.NoError(t, err)
	require.Equal(t, numberOfVouchersForUser, len(vouchersForUser.Vouchers))

	vouchersForOtherUser, err := GetVouchersForUser(testutils.GetAuthenticatedContext(otherUserPubKey), &GetVouchersForUserParams{UserPubKey: otherUserPubKey})
	require.NoError(t, err)
	require.Equal(t, numberOfVouchersForOtherUser, len(vouchersForOtherUser.Vouchers))
}

func TestGetVouchersForForNonExistingUser(t *testing.T) {
	testutils.EnsureExclusiveDatabaseAccess(t)
	testutils.ClearAllDBs()

	userPubKey, _ := testutils.GenerateKeys()
	all, err := GetAllVouchers(testutils.GetAuthenticatedContext(userPubKey), &GetAllVouchersParams{})
	require.NoError(t, err)
	require.NotNil(t, all.Vouchers)
	require.Equal(t, 0, len(all.Vouchers))

	usersVouchers, err := GetVouchersForUser(testutils.GetAuthenticatedContext(userPubKey), &GetVouchersForUserParams{UserPubKey: userPubKey})
	require.NoError(t, err)
	require.NotNil(t, usersVouchers.Vouchers)
	require.Equal(t, 0, len(usersVouchers.Vouchers))
}

func TestInvalidateVoucher(t *testing.T) {
	testutils.EnsureExclusiveDatabaseAccess(t)
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	ownerPubKey, _ := testutils.GenerateKeys()
	otherUser, _ := testutils.GenerateKeys()

	orgSigningPubKey, _ := testutils.GenerateKeys()
	orgEncryptionPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:               testOrganizationId,
		Name:             testOrganizationId,
		SigningPubKey:    orgSigningPubKey,
		EncryptionPubKey: orgEncryptionPubKey,
	})
	require.NoError(t, err)

	voucherDefinition, err := CreateVoucherDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &CreateVoucherDefinitionParams{
		OrganizationID: testOrganizationId,
		Name:           "My voucher",
		PictureURL:     "https://whatever.com/pic.jpeg",
	})

	testTable := []struct {
		name             string
		useRealVoucherID bool
		errorCode        errs.ErrCode
		uid              string
	}{
		{
			name:             "Happy path",
			useRealVoucherID: true,
			errorCode:        errs.OK,
			uid:              ownerPubKey,
		},
		{
			name:             "Different user",
			useRealVoucherID: true,
			errorCode:        errs.PermissionDenied,
			uid:              otherUser,
		},
		{
			name:             "Admin user",
			useRealVoucherID: true,
			errorCode:        errs.OK,
			uid:              testutils.AdminPubKey,
		},
		{
			name:             "Not found",
			useRealVoucherID: false,
			errorCode:        errs.NotFound,
			uid:              ownerPubKey,
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, testutils.ClearDB(depositDB, "voucher"))

			tx, err := depositDB.Begin(context.Background())
			require.NoError(t, err)
			mintedVoucherId, err := mintVoucher(testutils.GetAuthenticatedContext(testutils.AdminPubKey), tx, voucherDefinition, ownerPubKey)
			require.NoError(t, err)
			require.NoError(t, tx.Commit())

			ctx := testutils.GetAuthenticatedContext(test.uid)

			voucherID := mintedVoucherId
			if !test.useRealVoucherID {
				voucherID = "does not exist"
			}
			err = InvalidateVoucher(ctx, &InvalidateVoucherParams{
				VoucherID: voucherID,
			})
			if test.errorCode == errs.OK {
				require.NoError(t, err)

				dbVoucherRes, err := GetVoucher(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetVoucherParams{VoucherID: mintedVoucherId})
				require.NoError(t, err)
				require.True(t, dbVoucherRes.Voucher.Invalidated)
			} else {
				require.Error(t, err)
				require.Equal(t, test.errorCode, err.(*errs.Error).Code)

				dbVoucherRes, err := GetVoucher(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetVoucherParams{VoucherID: mintedVoucherId})
				require.NoError(t, err)
				require.False(t, dbVoucherRes.Voucher.Invalidated)
			}
		})
	}
}
