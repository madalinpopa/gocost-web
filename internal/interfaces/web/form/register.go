package form

type RegisterForm struct {
	Email    string `form:"email"`
	Username string `form:"username"`
	Password string `form:"password"`
	Base     `form:"-"`
}

func (f *RegisterForm) Validate() {
	f.CheckField(NotBlank(f.Email),
		"email",
		"this field is required",
	)
	f.CheckField(Matches(f.Email, MailRX),
		"email",
		"please enter a valid e-mail address",
	)
	f.CheckField(NotBlank(f.Username),
		"username",
		"this field is required",
	)
	f.CheckField(MinChars(f.Username, 3),
		"username",
		"username must be at least 3 characters long",
	)
	f.CheckField(MaxChars(f.Username, 30),
		"username",
		"username must be at most 30 characters long",
	)
	f.CheckField(UsernameCharsOnly(f.Username),
		"username",
		"username can only contain letters, numbers and underscores",
	)
	f.CheckField(NotBlank(f.Password),
		"password",
		"this field is required",
	)
	f.CheckField(MinChars(f.Password, 8),
		"password",
		"password must be at least 8 characters long",
	)
}
