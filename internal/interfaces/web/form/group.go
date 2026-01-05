package form

type CreateGroupForm struct {
	Name        string `form:"group-name"`
	Description string `form:"group-desc"`
	Order       int    `form:"group-order"`
	Base        `form:"-"`
}

func (f *CreateGroupForm) Validate() {
	f.CheckField(NotBlank(f.Name),
		"group-name",
		"this field is required",
	)
	f.CheckField(MaxChars(f.Name, 100),
		"group-name",
		"name must be at most 100 characters long",
	)
	f.CheckField(MaxChars(f.Description, 1000),
		"group-desc",
		"description must be at most 1000 characters long",
	)
	f.CheckField(f.Order >= 0,
		"group-order",
		"order must be non-negative",
	)
}

type UpdateGroupForm struct {
	ID          string `form:"group-id"`
	Name        string `form:"edit-group-name"`
	Description string `form:"edit-group-desc"`
	Order       int    `form:"edit-group-order"`
	Base        `form:"-"`
}

func (f *UpdateGroupForm) Validate() {
	f.CheckField(NotBlank(f.ID),
		"group-id",
		"group ID is required",
	)
	f.CheckField(NotBlank(f.Name),
		"edit-group-name",
		"this field is required",
	)
	f.CheckField(MaxChars(f.Name, 100),
		"edit-group-name",
		"name must be at most 100 characters long",
	)
	f.CheckField(MaxChars(f.Description, 1000),
		"edit-group-desc",
		"description must be at most 1000 characters long",
	)
	f.CheckField(f.Order >= 0,
		"edit-group-order",
		"order must be non-negative",
	)
}
