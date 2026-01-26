package form

import "strconv"

type CreateIncomeForm struct {
	Amount       string `form:"income-amount"`
	Description  string `form:"income-desc"`
	CurrentMonth string `form:"current-month"`
	Base         `form:"-"`
}

func (f *CreateIncomeForm) ParsedAmount() float64 {
	val, _ := strconv.ParseFloat(f.Amount, 64)
	return val
}

func (f *CreateIncomeForm) Validate() {
	if !ValidFloat(f.Amount) {
		f.AddFieldError("income-amount", "amount must be a number")
	} else {
		f.CheckField(PositiveFloat(f.ParsedAmount()),
			"income-amount",
			"amount must be greater than 0",
		)
	}
	f.CheckField(NotBlank(f.Description),
		"income-desc",
		"this field is required",
	)
	f.CheckField(MaxChars(f.Description, 100),
		"income-desc",
		"description must be at most 100 characters long",
	)
}
