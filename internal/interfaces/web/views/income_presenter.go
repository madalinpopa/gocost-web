package views

import (
	"fmt"

	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/madalinpopa/gocost-web/internal/usecase"
)

type IncomeView struct {
	ID            string
	Source        string
	ReceivedAt    string
	AmountDisplay string
}

type IncomeListPresenter struct {
	currency string
}

func NewIncomeListPresenter(currency string) *IncomeListPresenter {
	return &IncomeListPresenter{currency: currency}
}

func (p *IncomeListPresenter) Present(incomes []*usecase.IncomeResponse) []IncomeView {
	views := make([]IncomeView, 0, len(incomes))
	for _, inc := range incomes {
		if inc == nil {
			continue
		}

		views = append(views, IncomeView{
			ID:            inc.ID,
			Source:        inc.Source,
			ReceivedAt:    inc.ReceivedAt.Format(dateLayout),
			AmountDisplay: p.formatAmount(inc.Amount),
		})
	}

	return views
}

func (p *IncomeListPresenter) formatAmount(amount float64) string {
	m, err := money.NewFromFloat(amount, p.currency)
	if err == nil {
		return m.Display()
	}

	if p.currency == "" {
		return fmt.Sprintf("%.2f", amount)
	}

	return fmt.Sprintf("%s %.2f", p.currency, amount)
}
