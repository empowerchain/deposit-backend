package organization

import (
	"context"
	"strconv"
	"testing"

	"encore.app/admin"
	"encore.app/commons/testutils"
	"github.com/stretchr/testify/require"
)

func TestGetUsersFromOrganization(t *testing.T) {
	testutils.EnsureExclusiveDatabaseAccess(t)
	fakeOrgName := "testing"

	require.NoError(t, admin.InsertTestData(context.Background()))
	testutils.ClearAllDBs()

	users, err := GetUsersFromOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetUsersFromOrganizationParams{OrganizationId: "1"})
	require.NoError(t, err)

	require.Equal(t, len(users.Users), 0)

	signingPubKey, _ := testutils.GenerateKeys()
	encryptionPubKey, _ := testutils.GenerateKeys()

	_, err = CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &CreateOrgParams{
		ID:               "000",
		Name:             fakeOrgName,
		SigningPubKey:    signingPubKey,
		EncryptionPubKey: encryptionPubKey,
	})
	require.NoError(t, err)

	users, err = GetUsersFromOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetUsersFromOrganizationParams{OrganizationId: "000"})
	require.NoError(t, err)

	require.Equal(t, len(users.Users), 0)

	createUser, err := PostUserToOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &PostUserToOrganizationParams{OrganizationPubKey: "000", UserPubKey: "user", UserInformation: "content"})
	require.NoError(t, err)
	require.Equal(t, createUser.Successful, "post")

	newUser, err := GetUsersFromOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetUsersFromOrganizationParams{OrganizationId: "000"})
	require.NoError(t, err)

	require.Equal(t, newUser.OrganizationId, "000")
	require.Equal(t, len(newUser.Users), 1)
	require.Equal(t, newUser.Users[0].UserId, "user")
	require.Equal(t, newUser.Users[0].Content, "content")
}

func TestPostUserToOrganization(t *testing.T) {
	testutils.EnsureExclusiveDatabaseAccess(t)
	numberOfOrgs := 2
	orgName := "name"

	require.NoError(t, admin.InsertTestData(context.Background()))
	testutils.ClearAllDBs()

	for i := 0; i < numberOfOrgs; i++ {
		signingPubKey, _ := testutils.GenerateKeys()
		encryptionPubKey, _ := testutils.GenerateKeys()

		_, err := CreateOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &CreateOrgParams{
			ID:               strconv.Itoa(i),
			Name:             orgName,
			SigningPubKey:    signingPubKey,
			EncryptionPubKey: encryptionPubKey,
		})
		require.NoError(t, err)
	}

	allOrgs, err := GetAllOrganizations(testutils.GetAuthenticatedContext(testutils.AdminPubKey))
	require.NoError(t, err)
	require.Equal(t, numberOfOrgs, len(allOrgs.Organizations))

	for i, org := range allOrgs.Organizations {
		require.Equal(t, strconv.Itoa(i), org.ID)
		require.Equal(t, orgName, org.Name)
		require.NotEqual(t, "", org.SigningPubKey)
	}
	users, err := GetUsersFromOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetUsersFromOrganizationParams{OrganizationId: "1"})
	require.NoError(t, err)

	require.Equal(t, len(users.Users), 0)

	newUser, err := PostUserToOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &PostUserToOrganizationParams{OrganizationPubKey: "1", UserPubKey: "a", UserInformation: "content"})
	require.NoError(t, err)

	require.Equal(t, newUser.Successful, "post")

	getUser, err := GetUsersFromOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetUsersFromOrganizationParams{OrganizationId: "1"})
	require.NoError(t, err)
	require.Equal(t, len(getUser.Users), 1)

	updateUser, err := PostUserToOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &PostUserToOrganizationParams{OrganizationPubKey: "1", UserPubKey: "a", UserInformation: "updated content"})
	require.NoError(t, err)

	require.Equal(t, updateUser.Successful, "update")

	getUser, err = GetUsersFromOrganization(testutils.GetAuthenticatedContext(testutils.AdminPubKey), &GetUsersFromOrganizationParams{OrganizationId: "1"})
	require.NoError(t, err)
	require.Equal(t, len(getUser.Users), 1)
}
