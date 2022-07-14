package voucher

import (
	"context"
	"encore.app/admin"
	"encore.app/commons/testutils"
	"encore.app/organization"
	"encore.dev/beta/errs"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMintVoucher(t *testing.T) {
	testutils.ClearAllDBs()
	require.NoError(t, admin.InsertTestData(context.Background()))

	ownerPubKey, _ := testutils.GenerateKeys()

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
		name      string
		params    MintVoucherParams
		errorCode errs.ErrCode
		uid       string
	}{
		{
			name: "Happy path",
			params: MintVoucherParams{
				VoucherDefinitionID: voucherDefinition.ID,
				PubKey:              ownerPubKey,
			},
			errorCode: errs.OK,
			uid:       testutils.AdminPubKey,
		},
		{
			name: "Voucher def not found",
			params: MintVoucherParams{
				VoucherDefinitionID: "this does not exist",
				PubKey:              ownerPubKey,
			},
			errorCode: errs.NotFound,
			uid:       testutils.AdminPubKey,
		},
		{
			name: "unauthenticated",
			params: MintVoucherParams{
				VoucherDefinitionID: voucherDefinition.ID,
				PubKey:              ownerPubKey,
			},
			errorCode: errs.Unauthenticated,
			uid:       "",
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, testutils.ClearDB(voucherDB, "voucher"))

			var ctx context.Context
			if test.uid != "" {
				ctx = testutils.GetAuthenticatedContext(test.uid)
			} else {
				ctx = context.Background()
			}
			resp, err := MintVoucher(ctx, &test.params)
			if test.errorCode == errs.OK {
				require.NoError(t, err)
				require.NotEqual(t, "", resp.ID)
				require.Equal(t, test.params.VoucherDefinitionID, resp.VoucherDefinitionID)
				require.Equal(t, test.params.PubKey, resp.OwnerPubKey)

				allVouchers, err := GetAllVouchers(testutils.GetAuthenticatedContext(testutils.AdminPubKey))
				require.NoError(t, err)
				require.Equal(t, 1, len(allVouchers.Vouchers))

				dbVoucher, err := GetVoucher(ctx, &GetVoucherParams{VoucherID: resp.ID})
				require.NoError(t, err)
				require.Equal(t, resp.ID, dbVoucher.ID)
				require.Equal(t, test.params.VoucherDefinitionID, dbVoucher.VoucherDefinitionID)
				require.Equal(t, test.params.PubKey, dbVoucher.OwnerPubKey)
			} else {
				require.Error(t, err)
				require.Equal(t, test.errorCode, err.(*errs.Error).Code)

				allVouchers, err := GetAllVouchers(testutils.GetAuthenticatedContext(testutils.AdminPubKey))
				require.NoError(t, err)
				require.Equal(t, 0, len(allVouchers.Vouchers))
			}
		})
	}
}

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
			require.NoError(t, testutils.ClearDB(voucherDB, "voucher"))

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
