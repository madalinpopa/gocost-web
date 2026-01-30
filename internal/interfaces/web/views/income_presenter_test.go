package views

import (
	"testing"
	"time"

	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/stretchr/testify/assert"
)

func TestIncomeListPresenter_Present_FormatsIncome(t *testing.T) {
	presenter := NewIncomeListPresenter("USD")

	incomes := []*usecase.IncomeResponse{
		{
			ID:          "inc-1",
			Source:      "Salary",
			AmountCents: 10050,
			Currency:    "USD",
			ReceivedAt:  time.Date(2024, time.February, 3, 0, 0, 0, 0, time.UTC),
		},
	}

	views := presenter.Present(incomes)

	assert.Len(t, views, 1)
	assert.Equal(t, "inc-1", views[0].ID)
	assert.Equal(t, "Salary", views[0].Source)
	assert.Equal(t, "2024-02-03", views[0].ReceivedAt)
	assert.Equal(t, "$100.50", views[0].AmountDisplay)
}

func TestIncomeListPresenter_Present_SkipsNil(t *testing.T) {
	presenter := NewIncomeListPresenter("USD")

	incomes := []*usecase.IncomeResponse{
		nil,
		{
			ID:          "inc-2",
			Source:      "Bonus",
			AmountCents: 5000,
			Currency:    "USD",
			ReceivedAt:  time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	views := presenter.Present(incomes)

	assert.Len(t, views, 1)
	assert.Equal(t, "inc-2", views[0].ID)
}

func TestIncomeListPresenter_Present_FallbackFormatting(t *testing.T) {
	t.Run("empty currency", func(t *testing.T) {
		presenter := NewIncomeListPresenter("")

		incomes := []*usecase.IncomeResponse{
			{
				ID:          "inc-3",
				Source:      "Misc",
				AmountCents: 1230,
				ReceivedAt:  time.Date(2024, time.April, 2, 0, 0, 0, 0, time.UTC),
			},
		}

		views := presenter.Present(incomes)

		assert.Len(t, views, 1)
		assert.Equal(t, "12.30", views[0].AmountDisplay)
	})

	t.Run("invalid currency", func(t *testing.T) {
		presenter := NewIncomeListPresenter("BAD")

		incomes := []*usecase.IncomeResponse{
			{
				ID:          "inc-4",
				Source:      "Misc",
				AmountCents: 1230,
				ReceivedAt:  time.Date(2024, time.April, 2, 0, 0, 0, 0, time.UTC),
			},
		}

		views := presenter.Present(incomes)

		assert.Len(t, views, 1)
		assert.Equal(t, "BAD 12.30", views[0].AmountDisplay)
	})
}
