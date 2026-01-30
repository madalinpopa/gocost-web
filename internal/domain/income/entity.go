package income

import (
	"time"

	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
)

type ID = identifier.ID

type Income struct {
	ID         ID
	UserID     ID
	Amount     money.Money
	Source     SourceVO
	ReceivedAt time.Time
}

func NewIncome(id ID, userID ID, amount money.Money, source SourceVO, receivedAt time.Time) (*Income, error) {
	isPositive, err := amount.IsPositive()
	if err != nil || !isPositive {
		return nil, ErrInvalidAmount
	}

	return &Income{
		ID:         id,
		UserID:     userID,
		Amount:     amount,
		Source:     source,
		ReceivedAt: receivedAt,
	}, nil
}
