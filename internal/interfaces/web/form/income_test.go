package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateIncomeForm_Validate(t *testing.T) {
	tests := []struct {
		name       string
		form       CreateIncomeForm
		wantValid  bool
		wantErrors map[string]string
	}{
		{
			name: "valid form",
			form: CreateIncomeForm{
				Amount:      "100.50",
				Description: "Salary",
			},
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name: "invalid amount",
			form: CreateIncomeForm{
				Amount:      "0",
				Description: "Salary",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"income-amount": "amount must be greater than 0",
			},
		},
		{
			name: "invalid amount - not a number",
			form: CreateIncomeForm{
				Amount:      "abc",
				Description: "Salary",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"income-amount": "amount must be a number",
			},
		},
		{
			name: "missing description",
			form: CreateIncomeForm{
				Amount:      "100",
				Description: "",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"income-desc": "this field is required",
			},
		},
		{
			name: "description too long",
			form: CreateIncomeForm{
				Amount:      "100",
				Description: "this is a very long description that definitely exceeds the one hundred character limit defined in the validation rules for this field",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"income-desc": "description must be at most 100 characters long",
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
