package form

type LoginForm struct {
	Email    string `form:"email"`
	Password string `form:"password"`
	Base     `form:"-"`
}

func (f *LoginForm) Validate() {
	f.CheckField(NotBlank(f.Email),
		"email",
		"this field is required",
	)
	f.CheckField(Matches(f.Email, MailRX),
		"email",
		"please enter a valid e-mail address",
	)
	f.CheckField(NotBlank(f.Password),
		"password",
		"this field is required",
	)
}
