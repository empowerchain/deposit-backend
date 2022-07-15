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

func TestGetAllVouchers(t *testing.T) {
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	numberOfVouchers := 3

	organizationPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:     testOrganizationId,
		Name:   testOrganizationId,
		PubKey: organizationPubKey,
	})
	require.NoError(t, err)

	voucherDefinition, err := CreateVoucherDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &CreateVoucherDefinitionParams{
		OrganizationID: testOrganizationId,
		Name:           "My voucher",
		PictureURL:     "https://whatever.com/pic.jpeg",
	})

	for i := 0; i < numberOfVouchers; i++ {
		pubKey, _ := testutils.GenerateKeys()

		_, err := MintVoucher(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &MintVoucherParams{
			VoucherDefinitionID: voucherDefinition.ID,
			PubKey:              pubKey,
		})
		require.NoError(t, err)
	}

	allVouchers, err := GetAllVouchers(testutils.GetAuthenticatedContext(testutils.AdminPubKey))
	require.NoError(t, err)
	require.Equal(t, numberOfVouchers, len(allVouchers.Vouchers))

	for _, voucher := range allVouchers.Vouchers {
		require.Equal(t, voucherDefinition.ID, voucher.VoucherDefinitionID)
		require.NotEqual(t, "", voucher.ID)
		require.NotEqual(t, "pictureURL", voucher.OwnerPubKey)
	}
}

func TestGetAllVouchersForUser(t *testing.T) {
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	userPubKey, _ := testutils.GenerateKeys()
	numberOfVouchersForUser := 3
	otherUserPubKey, _ := testutils.GenerateKeys()
	numberOfVouchersForOtherUser := 7

	organizationPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:     testOrganizationId,
		Name:   testOrganizationId,
		PubKey: organizationPubKey,
	})
	require.NoError(t, err)

	voucherDefinition, err := CreateVoucherDefinition(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &CreateVoucherDefinitionParams{
		OrganizationID: testOrganizationId,
		Name:           "My voucher",
		PictureURL:     "https://whatever.com/pic.jpeg",
	})

	for i := 0; i < numberOfVouchersForUser; i++ {
		_, err := MintVoucher(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &MintVoucherParams{
			VoucherDefinitionID: voucherDefinition.ID,
			PubKey:              userPubKey,
		})
		require.NoError(t, err)
	}

	for i := 0; i < numberOfVouchersForOtherUser; i++ {
		_, err := MintVoucher(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &MintVoucherParams{
			VoucherDefinitionID: voucherDefinition.ID,
			PubKey:              otherUserPubKey,
		})
		require.NoError(t, err)
	}

	allVouchers, err := GetAllVouchers(testutils.GetAuthenticatedContext(testutils.AdminPubKey))
	require.NoError(t, err)
	require.Equal(t, numberOfVouchersForUser+numberOfVouchersForOtherUser, len(allVouchers.Vouchers))

	vouchersForUser, err := GetVouchersForUser(testutils.GetAuthenticatedContext(userPubKey), &GetVouchersForUserParams{UserPubKey: userPubKey})
	require.NoError(t, err)
	require.Equal(t, numberOfVouchersForUser, len(vouchersForUser.Vouchers))

	vouchersForOtherUser, err := GetVouchersForUser(testutils.GetAuthenticatedContext(otherUserPubKey), &GetVouchersForUserParams{UserPubKey: otherUserPubKey})
	require.NoError(t, err)
	require.Equal(t, numberOfVouchersForOtherUser, len(vouchersForOtherUser.Vouchers))
}

func TestInvalidateVoucher(t *testing.T) {
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	ownerPubKey, _ := testutils.GenerateKeys()
	otherUser, _ := testutils.GenerateKeys()

	organizationPubKey, _ := testutils.GenerateKeys()
	_, err := organization.CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &organization.CreateOrgParams{
		ID:     testOrganizationId,
		Name:   testOrganizationId,
		PubKey: organizationPubKey,
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
			name:             "Unauthenticated",
			useRealVoucherID: true,
			errorCode:        errs.Unauthenticated,
			uid:              "",
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

			voucher, err := MintVoucher(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &MintVoucherParams{
				VoucherDefinitionID: voucherDefinition.ID,
				PubKey:              ownerPubKey,
			})
			require.NoError(t, err)

			ctx := testutils.GetAuthenticatedContext(test.uid)
			if test.uid == "" {
				ctx = context.Background()
			}

			voucherID := voucher.ID
			if !test.useRealVoucherID {
				voucherID = "does not exist"
			}
			err = InvalidateVoucher(ctx, &InvalidateVoucherParams{
				VoucherID: voucherID,
			})
			if test.errorCode == errs.OK {
				require.NoError(t, err)

				dbVoucher, err := GetVoucher(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetVoucherParams{VoucherID: voucher.ID})
				require.NoError(t, err)
				require.True(t, dbVoucher.Invalidated)
			} else {
				require.Error(t, err)
				require.Equal(t, test.errorCode, err.(*errs.Error).Code)

				dbVoucher, err := GetVoucher(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetVoucherParams{VoucherID: voucher.ID})
				require.NoError(t, err)
				require.False(t, dbVoucher.Invalidated)
			}
		})
	}
}
