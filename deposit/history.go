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
	EventType       string    `json:"eventType"`
	EventTime       time.Time `json:"eventTime"`
	UnitNameIn      string    `json:"unitNameIn"`
	NumberOfUnitsIn float64   `json:"numberOfUnitsIn"`
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
		rewards, err := getRewards(ctx, &d)
		if err != nil {
			return nil, err
		}
		for _, r := range rewards {

			events = append(events, Event{
				EventType:       EventTypeDeposit,
				EventTime:       d.CreatedAt,
				UnitNameIn:      "Voucher", //TODO: Make better
				NumberOfUnitsIn: r.Amount,
			})
		}
	}

	if events == nil {
		events = []Event{}
	}
	return &GetHistoryResponse{Events: events}, nil
}
