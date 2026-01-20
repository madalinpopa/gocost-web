package form

type CreateCategoryForm struct {
	GroupID     string  `form:"group-id"`
	Name        string  `form:"category-name"`
	Description string  `form:"category-desc"`
	Type        string  `form:"type"`
	StartMonth  string  `form:"category-start"`
	EndMonth    string  `form:"category-end"`
	Budget      float64 `form:"category-budget"`
	Base        `form:"-"`
}

func (f *CreateCategoryForm) Validate() {
	f.CheckField(NotBlank(f.GroupID),
		"group-id",
		"group ID is required",
	)
	f.CheckField(NotBlank(f.Name),
		"category-name",
		"this field is required",
	)
	f.CheckField(MaxChars(f.Name, 100),
		"category-name",
		"name must be at most 100 characters long",
	)
	f.CheckField(MaxChars(f.Description, 1000),
		"category-desc",
		"description must be at most 1000 characters long",
	)
	f.CheckField(PermittedValue(f.Type, "monthly", "recurrent"),
		"type",
		"invalid category type",
	)
	f.CheckField(f.Budget >= 0,
		"category-budget",
		"budget must be zero or positive",
	)
	if !NotBlank(f.StartMonth) {
		f.AddFieldError("category-start", "this field is required")
	} else {
		f.CheckField(ValidMonthString(f.StartMonth),
			"category-start",
			"invalid month format",
		)
	}

	if f.EndMonth != "" {
		f.CheckField(ValidMonthString(f.EndMonth),
			"category-end",
			"invalid month format",
		)
		if f.Type == "recurrent" && ValidMonthString(f.StartMonth) && ValidMonthString(f.EndMonth) {
			f.CheckField(f.EndMonth >= f.StartMonth,
				"category-end",
				"end month must be after start month",
			)
		}
	}
}

type UpdateCategoryForm struct {
	ID           string  `form:"category-id"`
	GroupID      string  `form:"group-id"`
	Name         string  `form:"edit-name"`
	Description  string  `form:"edit-desc"`
	Type         string  `form:"type"`
	StartMonth   string  `form:"edit-start"`
	EndMonth     string  `form:"edit-end"`
	CurrentMonth string  `form:"current-month"`
	Budget       float64 `form:"edit-budget"`
	Base         `form:"-"`
}

func (f *UpdateCategoryForm) Validate() {
	f.CheckField(NotBlank(f.ID),
		"category-id",
		"category ID is required",
	)
	f.CheckField(NotBlank(f.GroupID),
		"group-id",
		"group ID is required",
	)
	f.CheckField(NotBlank(f.Name),
		"edit-name",
		"this field is required",
	)
	f.CheckField(MaxChars(f.Name, 100),
		"edit-name",
		"name must be at most 100 characters long",
	)
	f.CheckField(MaxChars(f.Description, 1000),
		"edit-desc",
		"description must be at most 1000 characters long",
	)
	f.CheckField(PermittedValue(f.Type, "monthly", "recurrent"),
		"type",
		"invalid category type",
	)
	f.CheckField(f.Budget >= 0,
		"edit-budget",
		"budget must be zero or positive",
	)
	if !NotBlank(f.StartMonth) {
		f.AddFieldError("edit-start", "this field is required")
	} else {
		f.CheckField(ValidMonthString(f.StartMonth),
			"edit-start",
			"invalid month format",
		)
	}

	if f.EndMonth != "" {
		f.CheckField(ValidMonthString(f.EndMonth),
			"edit-end",
			"invalid month format",
		)
		if f.Type == "recurrent" && ValidMonthString(f.StartMonth) && ValidMonthString(f.EndMonth) {
			f.CheckField(f.EndMonth >= f.StartMonth,
				"edit-end",
				"end month must be after start month",
			)
		}
	}
}
