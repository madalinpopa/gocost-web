package income

import (
	"time"

	"github.com/madalinpopa/gocost-web/internal/shared/identifier"
	"github.com/madalinpopa/gocost-web/internal/shared/money"
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
	if amount.IsZero() || !amount.IsPositive() {
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
