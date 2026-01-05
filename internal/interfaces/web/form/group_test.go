package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateGroupForm_Validate(t *testing.T) {
	tests := []struct {
		name       string
		form       CreateGroupForm
		wantValid  bool
		wantErrors map[string]string
	}{
		{
			name: "valid form",
			form: CreateGroupForm{
				Name:        "Housing",
				Description: "Rent and utilities",
			},
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name: "missing name",
			form: CreateGroupForm{
				Name:        "",
				Description: "Rent and utilities",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"group-name": "this field is required",
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

func TestUpdateGroupForm_Validate(t *testing.T) {
	tests := []struct {
		name       string
		form       UpdateGroupForm
		wantValid  bool
		wantErrors map[string]string
	}{
		{
			name: "valid form",
			form: UpdateGroupForm{
				ID:          "123",
				Name:        "Housing",
				Description: "Rent and utilities",
			},
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name: "missing id",
			form: UpdateGroupForm{
				ID:          "",
				Name:        "Housing",
				Description: "Rent and utilities",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"group-id": "group ID is required",
			},
		},
		{
			name: "missing name",
			form: UpdateGroupForm{
				ID:          "123",
				Name:        "",
				Description: "Rent and utilities",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"edit-group-name": "this field is required",
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
