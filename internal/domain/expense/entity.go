package expense

import (
	"time"

	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
)

type ID = identifier.ID

type Expense struct {
	ID          ID
	CategoryID  ID
	Amount      money.Money
	Description ExpenseDescriptionVO
	SpentAt     time.Time
	Payment     PaymentStatus
}

func NewExpense(id ID, categoryID ID, amount money.Money, description ExpenseDescriptionVO, spentAt time.Time, payment PaymentStatus) (*Expense, error) {
	if amount.IsZero() || !amount.IsPositive() {
		return nil, ErrInvalidAmount
	}

	return &Expense{
		ID:          id,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		SpentAt:     spentAt,
		Payment:     payment,
	}, nil
}
