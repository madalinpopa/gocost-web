package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateExpenseForm_Validate(t *testing.T) {
	tests := []struct {
		name       string
		form       CreateExpenseForm
		wantValid  bool
		wantErrors map[string]string
	}{
		{
			name: "valid unpaid expense",
			form: CreateExpenseForm{
				CategoryID:    "cat-123",
				Amount:        50.00,
				Month:         "2023-10",
				PaymentStatus: "unpaid",
			},
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name: "valid paid expense",
			form: CreateExpenseForm{
				CategoryID:    "cat-123",
				Amount:        50.00,
				Month:         "2023-10",
				PaymentStatus: "paid",
			},
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name: "invalid amount",
			form: CreateExpenseForm{
				CategoryID:    "cat-123",
				Amount:        -10.00,
				Month:         "2023-10",
				PaymentStatus: "unpaid",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"expense-amount": "amount must be greater than 0",
			},
		},
		{
			name: "invalid month format",
			form: CreateExpenseForm{
				CategoryID:    "cat-123",
				Amount:        10.00,
				Month:         "2023-10-27",
				PaymentStatus: "unpaid",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"month": "invalid month format",
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

func TestUpdateExpenseForm_Validate(t *testing.T) {
	tests := []struct {
		name       string
		form       UpdateExpenseForm
		wantValid  bool
		wantErrors map[string]string
	}{
		{
			name: "valid update",
			form: UpdateExpenseForm{
				ID:            "exp-123",
				CategoryID:    "cat-123",
				Amount:        75.00,
				PaymentStatus: "unpaid",
			},
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name: "missing expense ID",
			form: UpdateExpenseForm{
				CategoryID:    "cat-123",
				Amount:        75.00,
				PaymentStatus: "unpaid",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"expense-id": "expense ID is required",
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
