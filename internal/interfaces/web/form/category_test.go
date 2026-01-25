package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCategoryForm_Validate(t *testing.T) {
	tests := []struct {
		name       string
		form       CreateCategoryForm
		wantValid  bool
		wantErrors map[string]string
	}{
		{
			name: "valid monthly category",
			form: CreateCategoryForm{
				GroupID:    "123",
				Name:       "Rent",
				Type:       "monthly",
				StartMonth: "2023-10",
				Budget:     "100.00",
			},
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name: "valid recurrent category",
			form: CreateCategoryForm{
				GroupID:    "123",
				Name:       "Rent",
				Type:       "recurrent",
				StartMonth: "2023-01",
				EndMonth:   "2023-12",
				Budget:     "100.00",
			},
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name: "missing group id",
			form: CreateCategoryForm{
				Name:       "Rent",
				Type:       "monthly",
				StartMonth: "2023-10",
				Budget:     "100.00",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"group-id": "group ID is required",
			},
		},
		{
			name: "end month before start month",
			form: CreateCategoryForm{
				GroupID:    "123",
				Name:       "Rent",
				Type:       "recurrent",
				StartMonth: "2023-12",
				EndMonth:   "2023-01",
				Budget:     "100.00",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"category-end": "end month must be after start month",
			},
		},
		{
			name: "invalid budget - not a number",
			form: CreateCategoryForm{
				GroupID:    "123",
				Name:       "Rent",
				Type:       "monthly",
				StartMonth: "2023-10",
				Budget:     "abc",
			},
			wantValid: false,
			wantErrors: map[string]string{
				"category-budget": "budget must be a number",
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