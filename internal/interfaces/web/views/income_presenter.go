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
			AmountDisplay: p.formatAmount(inc.AmountCents, inc.Currency),
		})
	}

	return views
}

func (p *IncomeListPresenter) formatAmount(cents int64, currency string) string {
	displayCurrency := currency
	if displayCurrency == "" {
		displayCurrency = p.currency
	}

	if displayCurrency != "" {
		m, err := money.New(cents, displayCurrency)
		if err == nil {
			return m.Display()
		}
	}

	return formatCents(cents, displayCurrency)
}

func formatCents(cents int64, currency string) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}

	units := cents / 100
	fraction := cents % 100

	if currency == "" {
		return fmt.Sprintf("%s%d.%02d", sign, units, fraction)
	}

	return fmt.Sprintf("%s%s %d.%02d", sign, currency, units, fraction)
}
