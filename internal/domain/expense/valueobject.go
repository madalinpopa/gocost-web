package expense

import (
	"fmt"
	"time"
)

const (
	monthLayout = "2006-01"
)

type Month struct {
	value string
}

func NewMonth(year int, month time.Month) (Month, error) {
	if year < 1 || month < time.January || month > time.December {
		return Month{}, ErrInvalidMonth
	}

	return Month{value: fmt.Sprintf("%04d-%02d", year, month)}, nil
}

func ParseMonth(value string) (Month, error) {
	parsed, err := time.Parse(monthLayout, value)
	if err != nil {
		return Month{}, ErrInvalidMonth
	}

	return Month{value: parsed.Format(monthLayout)}, nil
}

func NewMonthFromTime(value time.Time) Month {
	return Month{value: value.Format(monthLayout)}
}

func (m Month) Value() string {
	return m.value
}

func (m Month) String() string {
	return m.value
}

func (m Month) Equals(other Month) bool {
	return m.value == other.value
}

func (m Month) IsZero() bool {
	return m.value == ""
}

type ExpenseDescriptionVO struct {
	value string
}

func NewExpenseDescriptionVO(value string) (ExpenseDescriptionVO, error) {
	if len(value) > 255 {
		return ExpenseDescriptionVO{}, ErrExpenseDescriptionTooLong
	}
	return ExpenseDescriptionVO{value: value}, nil
}

func (d ExpenseDescriptionVO) Value() string {
	return d.value
}

func (d ExpenseDescriptionVO) String() string {
	return d.value
}

func (d ExpenseDescriptionVO) Equals(other ExpenseDescriptionVO) bool {
	return d.value == other.value
}

type PaymentStatus struct {
	isPaid bool
	paidAt *time.Time
}

func NewPaymentStatus(isPaid bool, paidAt *time.Time) (PaymentStatus, error) {
	if isPaid {
		if paidAt == nil || paidAt.IsZero() {
			return PaymentStatus{}, ErrPaidAtRequired
		}
		paidAtCopy := *paidAt
		return PaymentStatus{isPaid: true, paidAt: &paidAtCopy}, nil
	}

	if paidAt != nil {
		return PaymentStatus{}, ErrPaidAtNotAllowed
	}

	return PaymentStatus{isPaid: false, paidAt: nil}, nil
}

func NewPaidStatus(paidAt time.Time) (PaymentStatus, error) {
	if paidAt.IsZero() {
		return PaymentStatus{}, ErrPaidAtRequired
	}
	return PaymentStatus{isPaid: true, paidAt: &paidAt}, nil
}

func NewUnpaidStatus() PaymentStatus {
	return PaymentStatus{isPaid: false, paidAt: nil}
}

func (p PaymentStatus) IsPaid() bool {
	return p.isPaid
}

func (p PaymentStatus) PaidAt() *time.Time {
	if p.paidAt == nil {
		return nil
	}
	paidAtCopy := *p.paidAt
	return &paidAtCopy
}
