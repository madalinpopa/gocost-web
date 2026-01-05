package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterForm_Validate(t *testing.T) {
	tests := []struct {
		name       string
		form       RegisterForm
		wantValid  bool
		wantErrors map[string]string
	}{
		{
			name: "valid form",
			form: RegisterForm{
				Email:    "user@example.com",
				Username: "test_user",
				Password: "supersecret",
			},
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name: "invalid email format",
			form: RegisterForm{
				Email:    "not-an-email",
				Username: "test_user",
				Password: "supersecret",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"email": "please enter a valid e-mail address",
			},
		},
		{
			name: "username too short",
			form: RegisterForm{
				Email:    "user@example.com",
				Username: "ab",
				Password: "supersecret",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"username": "username must be at least 3 characters long",
			},
		},
		{
			name: "username too long",
			form: RegisterForm{
				Email:    "user@example.com",
				Username: "this_username_is_longer_than_thirty_chars",
				Password: "supersecret",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"username": "username must be at most 30 characters long",
			},
		},
		{
			name: "password too short",
			form: RegisterForm{
				Email:    "user@example.com",
				Username: "test_user",
				Password: "short",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"password": "password must be at least 8 characters long",
			},
		},
		{
			name: "missing all fields",
			form: RegisterForm{
				Email:    "",
				Username: "",
				Password: "",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"email":    "please enter a valid e-mail address",
				"username": "username must be at least 3 characters long",
				"password": "password must be at least 8 characters long",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.form.Validate()

			assert.Equal(t, tt.wantValid, tt.form.IsValid())
			assert.Equal(t, tt.wantErrors, tt.form.FieldErrors)
		})
	}
}
