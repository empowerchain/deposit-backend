package deposit

import (
	"context"
	"encore.app/commons"
	"time"
)

const (
	EventTypeDeposit = "DEPOSIT"
)

type Event struct {
	EventType     string    `json:"eventType"`
	EventTime     time.Time `json:"eventTime"`
	UnitName      string    `json:"unitName"`
	NumberOfUnits float64   `json:"numberOfUnits"`
}

type GetHistoryParams struct {
	UserPubKey string `json:"userPubKey" validate:"required"`
}

type GetHistoryResponse struct {
	Events []Event `json:"events"`
}

// TODO: TEST THIS
//encore:api public method=POST
func GetHistory(ctx context.Context, params *GetHistoryParams) (*GetHistoryResponse, error) {
	if err := commons.Validate(params); err != nil {
		return nil, err
	}

	deposits, err := GetAllDeposits(ctx, &GetAllDepositsParams{
		UserPubKey: params.UserPubKey,
		Desc:       true,
	})
	if err != nil {
		return nil, err
	}

	var events []Event
	for _, d := range deposits.Deposits {
		for _, mb := range d.MassBalanceDeposits {

			events = append(events, Event{
				EventType:     EventTypeDeposit,
				EventTime:     d.CreatedAt,
				UnitName:      "Voucher", //TODO: Make better
				NumberOfUnits: mb.Amount,
			})
		}
	}

	if events == nil {
		events = []Event{}
	}
	return &GetHistoryResponse{Events: events}, nil
}
