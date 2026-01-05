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
				Amount:      100.50,
				Description: "Salary",
				Date:        "2023-10-27",
			},
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name: "invalid amount",
			form: CreateIncomeForm{
				Amount:      0,
				Description: "Salary",
				Date:        "2023-10-27",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"income-amount": "amount must be greater than 0",
			},
		},
		{
			name: "missing description",
			form: CreateIncomeForm{
				Amount:      100,
				Description: "",
				Date:        "2023-10-27",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"income-desc": "this field is required",
			},
		},
		{
			name: "description too long",
			form: CreateIncomeForm{
				Amount:      100,
				Description: "this is a very long description that definitely exceeds the one hundred character limit defined in the validation rules for this field",
				Date:        "2023-10-27",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"income-desc": "description must be at most 100 characters long",
			},
		},
		{
			name: "missing date",
			form: CreateIncomeForm{
				Amount:      100,
				Description: "Salary",
				Date:        "",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"income-date": "this field is required",
			},
		},
		{
			name: "invalid date format",
			form: CreateIncomeForm{
				Amount:      100,
				Description: "Salary",
				Date:        "invalid-date",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"income-date": "invalid date format",
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
