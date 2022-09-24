package stats

import (
	"context"

	"encore.app/commons"
	"encore.app/deposit"
	"encore.app/organization"
	"encore.app/scheme"
	"encore.dev/beta/errs"
	"golang.org/x/exp/maps"
)

type User struct {
	PubKey string `json:"user"`
}

type DepositDescription struct {
	Magnitude          int64             `json:"magnitude"`
	Amount             float64           `json:"amount"`
	MaterialDefinition map[string]string `json:"materialDefinition"`
}
type Stats struct {
	NumberOfAvailableVouchers int64                `json:"numberOfAvailableVouchers"`
	PlasticCollected          float64              `json:"plasticCollected"`
	NumberOfUsedVouchers      int64                `json:"numberOfUsedVouchers"`
	DepositAmounts            []DepositDescription `json:"depositAmounts"`
}

//encore:api public method=POST
func GetStats(ctx context.Context, params *User) (*Stats, error) {
	if err := commons.Validate(params); err != nil || params.PubKey == "" {
		return nil, &errs.Error{
			Code: errs.InvalidArgument,
		}
	}

	resp := &Stats{DepositAmounts: []DepositDescription{}}

	allVouchers, err := deposit.GetAllVouchers(ctx, &deposit.GetAllVouchersParams{PubKey: params.PubKey})
	if err != nil {
		return nil, err
	}

	for _, voucher := range allVouchers.Vouchers {
		if voucher.Voucher.Invalidated {
			resp.NumberOfUsedVouchers += 1
		} else {
			resp.NumberOfAvailableVouchers += 1
		}
	}

	allDeposits, err := deposit.GetAllDeposits(ctx, &deposit.GetAllDepositsParams{UserPubKey: params.PubKey})
	if err != nil {
		return nil, err
	}

	var materialsExisting = make(map[string]DepositDescription)

	var materials = []DepositDescription{}
	for _, register := range allDeposits.Deposits {
		var currentDepositDescription DepositDescription = DepositDescription{} // Deposit to add in list
		var masses = register.MassBalanceDeposits                               // List of massDeposits
		var materialKey string                                                  // plastic we're storing
		for _, mass := range masses {
			currentDepositDescription.MaterialDefinition = mass.ItemDefinition.MaterialDefinition // describes the material
			materialKey = maps.Values(mass.ItemDefinition.MaterialDefinition)[0]                  // Plastic stored
			currentDepositDescription.Amount += mass.Amount
			if mass.ItemDefinition.Magnitude == 0 {
				currentDepositDescription.Magnitude = 0
			} else {
				currentDepositDescription.Magnitude = 1
			}
			var updatedDepositDescription DepositDescription = materialsExisting[materialKey] // if it's empty np
			if _, usedMaterial := materialsExisting[materialKey]; usedMaterial {
				updatedDepositDescription.Amount += currentDepositDescription.Amount // If not, update amount
			} else {
				updatedDepositDescription = currentDepositDescription
			}
			materialsExisting[materialKey] = updatedDepositDescription
		}
	}
	for _, material := range materialsExisting {
		if material.Magnitude == 0 {
			resp.PlasticCollected += material.Amount
		}
		materials = append(materials, material)
	}

	resp.DepositAmounts = materials

	return resp, nil
}

type OrganizationData struct {
	ID   string `json:"organizationId"`
	Name string `json:"organizationName"`
}

type Organizations struct {
	DepositOrgsForUser []OrganizationData `json:"depositOrgsForUser"`
}

//encore:api public method=POST
func GetOrganizationsByUser(ctx context.Context, params *User) (*Organizations, error) {
	if err := commons.Validate(params); err != nil || params.PubKey == "" {
		return nil, &errs.Error{
			Code: errs.InvalidArgument,
		}
	}

	var resp = &Organizations{DepositOrgsForUser: []OrganizationData{}}

	allDeposits, _ := deposit.GetAllDeposits(ctx, &deposit.GetAllDepositsParams{UserPubKey: params.PubKey})

	var registeredOrganizations = make(map[string]string)

	for _, deposit := range allDeposits.Deposits {
		depositData, _ := scheme.GetScheme(ctx, &scheme.GetSchemeParams{SchemeID: deposit.SchemeID})
		organizationId := depositData.OrganizationID
		if _, organizationRegistered := registeredOrganizations[organizationId]; !organizationRegistered {
			organizationName, _ := organization.GetOrganization(ctx, &organization.GetOrganizationParams{ID: organizationId})
			registeredOrganizations[organizationId] = organizationName.Name // Updating map
			var currentOrganization = OrganizationData{Name: organizationName.Name, ID: organizationId}
			resp.DepositOrgsForUser = append(resp.DepositOrgsForUser, currentOrganization)
		}
	}

	return resp, nil
}
