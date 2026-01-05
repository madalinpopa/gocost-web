package form

type CreateExpenseForm struct {
	CategoryID    string  `form:"category-id"`
	Amount        float64 `form:"expense-amount"`
	Description   string  `form:"expense-desc"`
	Month         string  `form:"month"`
	PaymentStatus string  `form:"payment-status"`
	Base          `form:"-"`
}

func (f *CreateExpenseForm) Validate() {
	f.CheckField(NotBlank(f.CategoryID),
		"category-id",
		"category ID is required",
	)
	f.CheckField(PositiveFloat(f.Amount),
		"expense-amount",
		"amount must be greater than 0",
	)
	f.CheckField(MaxChars(f.Description, 255),
		"expense-desc",
		"description must be at most 255 characters long",
	)
	f.CheckField(ValidDateString(f.Month+"-01"),
		"month",
		"invalid month format",
	)
	f.CheckField(PermittedValue(f.PaymentStatus, "paid", "unpaid"),
		"payment-status",
		"invalid status",
	)
}

type UpdateExpenseForm struct {
	ID            string  `form:"expense-id"`
	CategoryID    string  `form:"category-id"`
	Amount        float64 `form:"edit-amount"`
	Description   string  `form:"edit-desc"`
	PaymentStatus string  `form:"payment-status"`
	Base          `form:"-"`
}

func (f *UpdateExpenseForm) Validate() {
	f.CheckField(NotBlank(f.ID),
		"expense-id",
		"expense ID is required",
	)
	f.CheckField(NotBlank(f.CategoryID),
		"category-id",
		"category ID is required",
	)
	f.CheckField(PositiveFloat(f.Amount),
		"edit-amount",
		"amount must be greater than 0",
	)
	f.CheckField(MaxChars(f.Description, 255),
		"edit-desc",
		"description must be at most 255 characters long",
	)
	f.CheckField(PermittedValue(f.PaymentStatus, "paid", "unpaid"),
		"payment-status",
		"invalid status",
	)
}
