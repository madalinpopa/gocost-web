package form

type CreateIncomeForm struct {
	Amount       float64 `form:"income-amount"`
	Description  string  `form:"income-desc"`
	Date         string  `form:"income-date"`
	CurrentMonth string  `form:"current-month"`
	Base         `form:"-"`
}

func (f *CreateIncomeForm) Validate() {
	f.CheckField(PositiveFloat(f.Amount),
		"income-amount",
		"amount must be greater than 0",
	)
	f.CheckField(NotBlank(f.Description),
		"income-desc",
		"this field is required",
	)
	f.CheckField(MaxChars(f.Description, 100),
		"income-desc",
		"description must be at most 100 characters long",
	)
	if !NotBlank(f.Date) {
		f.AddFieldError("income-date", "this field is required")
	} else {
		f.CheckField(ValidDateString(f.Date),
			"income-date",
			"invalid date format",
		)
	}
}
