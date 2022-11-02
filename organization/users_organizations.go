package organization

import (
	"context"

	"encore.app/commons"
	"encore.dev/storage/sqldb"
)

type PostUserToOrganizationParams struct {
	OrganizationPubKey string `json:"organization_pub_key"`
	UserPubKey         string `json:"user_pub_key"`
	UserInformation    string `json:"content"`
}

type UserPosted struct {
	Successful bool `json:"success"`
}

//encore:api public method=POST
func PostUserToOrganization(ctx context.Context, params *PostUserToOrganizationParams) (*UserPosted, error) {

	registerID := commons.GenerateID()
	organizationPubKey := params.OrganizationPubKey
	userPubKey := params.UserPubKey
	userInformation := params.UserInformation

	// TODO:
	// Use some proper ORM instead of raw SQL querie
	_, err := sqldb.Exec(ctx, "INSERT INTO user_organization (id, organization_pub_key, user_pub_key, content) VALUES ($1, $2, $3, $4);",
		registerID,
		organizationPubKey,
		userPubKey,
		userInformation)
	if err != nil {
		return nil, err
	}

	return &UserPosted{Successful: true}, nil
}

type GetUsersFromOrganizationParams struct {
	OrganizationId string `json:"organizationId"`
}

type UserFromOrganization struct {
	UserId  string `json:"user_pub_key"`
	Content string `json:"content"`
}

type UsersFromOrganization struct {
	OrganizationId string                 `json:"organization_pub_key"`
	Users          []UserFromOrganization `json:"users"`
}

//encore:api public method=GET
func GetUsersFromOrganization(ctx context.Context, params *GetUsersFromOrganizationParams) (*UsersFromOrganization, error) {

	organization := params.OrganizationId

	rows, err := sqldb.Query(ctx,
		"SELECT user_pub_key, content FROM user_organization WHERE organization_pub_key=$1",
		&organization)
	if err != nil {
		return nil, err
	}
	response := UsersFromOrganization{OrganizationId: organization}
	users := []UserFromOrganization{}

	for rows.Next() {
		user := UserFromOrganization{}
		if err := rows.Scan(&user.UserId, &user.Content); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	response.Users = users
	return &response, nil
}
