package tracking

import (
	"fmt"
	"time"
)

const (
	monthLayout = "2006-01"
)

type NameVO struct {
	value string
}

func NewNameVO(value string) (NameVO, error) {
	if value == "" {
		return NameVO{}, ErrEmptyName
	}
	if len(value) > 100 {
		return NameVO{}, ErrNameTooLong
	}
	return NameVO{value: value}, nil
}

func (n NameVO) Value() string {
	return n.value
}

func (n NameVO) String() string {
	return n.value
}

func (n NameVO) Equals(other NameVO) bool {
	return n.value == other.value
}

type DescriptionVO struct {
	value string
}

func NewDescriptionVO(value string) (DescriptionVO, error) {
	if len(value) > 1000 {
		return DescriptionVO{}, ErrDescriptionTooLong
	}
	return DescriptionVO{value: value}, nil
}

func (d DescriptionVO) Value() string {
	return d.value
}

func (d DescriptionVO) String() string {
	return d.value
}

func (d DescriptionVO) Equals(other DescriptionVO) bool {
	return d.value == other.value
}

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

func (m Month) Before(other Month) bool {
	return m.value < other.value
}

func (m Month) Previous() Month {
	t, _ := time.Parse(monthLayout, m.value)
	return NewMonthFromTime(t.AddDate(0, -1, 0))
}

func (m Month) Next() Month {
	t, _ := time.Parse(monthLayout, m.value)
	return NewMonthFromTime(t.AddDate(0, 1, 0))
}

type OrderVO struct {
	value int
}

func NewOrderVO(value int) (OrderVO, error) {
	if value < 0 {
		return OrderVO{}, ErrInvalidOrder
	}
	return OrderVO{value: value}, nil
}

func (o OrderVO) Value() int {
	return o.value
}

func (o OrderVO) Equals(other OrderVO) bool {
	return o.value == other.value
}
