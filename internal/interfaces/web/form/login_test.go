package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginForm_Validate(t *testing.T) {
	tests := []struct {
		name       string
		form       LoginForm
		wantValid  bool
		wantErrors map[string]string
	}{
		{
			name: "valid form",
			form: LoginForm{
				Email:    "user@example.com",
				Password: "secret",
			},
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name: "invalid email format",
			form: LoginForm{
				Email:    "not-an-email",
				Password: "secret",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"email": "please enter a valid e-mail address",
			},
		},
		{
			name: "missing password",
			form: LoginForm{
				Email:    "user@example.com",
				Password: "",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"password": "this field is required",
			},
		},
		{
			name: "missing email and password",
			form: LoginForm{
				Email:    "",
				Password: "",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"email":    "this field is required",
				"password": "this field is required",
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
