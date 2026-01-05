package usecase

import "time"

type IDRequest struct {
	ID string `json:"-"`
}

type EmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type UsernameRequest struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
}

type RegisterUserRequest struct {
	EmailRequest
	UsernameRequest
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	EmailOrUsername string `json:"email_or_username" validate:"required"`
	Password        string `json:"password" validate:"required"`
}

type LoginResponse struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type CreateIncomeRequest struct {
	Amount     float64   `json:"amount" validate:"required,gt=0"`
	Source     string    `json:"source" validate:"required,max=100"`
	ReceivedAt time.Time `json:"received_at" validate:"required"`
}

type UpdateIncomeRequest struct {
	ID         string    `json:"-"`
	Amount     float64   `json:"amount" validate:"required,gt=0"`
	Source     string    `json:"source" validate:"required,max=100"`
	ReceivedAt time.Time `json:"received_at" validate:"required"`
}

type IncomeResponse struct {
	ID         string    `json:"id"`
	Amount     float64   `json:"amount"`
	Source     string    `json:"source"`
	ReceivedAt time.Time `json:"received_at"`
}

type CreateGroupRequest struct {
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"max=255"`
	Order       int    `json:"order" validate:"min=0"`
}

type UpdateGroupRequest struct {
	ID          string `json:"-"`
	Name        string `json:"name" validate:"required,max=100"`
	Description string `json:"description" validate:"max=255"`
	Order       int    `json:"order" validate:"min=0"`
}

type CategoryResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	IsRecurrent bool    `json:"is_recurrent"`
	StartMonth  string  `json:"start_month"`
	EndMonth    string  `json:"end_month,omitempty"`
	Budget      float64 `json:"budget"`
}

type GroupResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Order       int                `json:"order"`
	Categories  []CategoryResponse `json:"categories"`
}

type CreateCategoryRequest struct {
	Name        string  `json:"name" validate:"required,max=100"`
	Description string  `json:"description" validate:"max=1000"`
	IsRecurrent bool    `json:"is_recurrent"`
	StartMonth  string  `json:"start_month" validate:"required"`
	EndMonth    string  `json:"end_month,omitempty"`
	Budget      float64 `json:"budget" validate:"min=0"`
}

type UpdateCategoryRequest struct {
	ID          string  `json:"-"`
	Name        string  `json:"name" validate:"required,max=100"`
	Description string  `json:"description" validate:"max=1000"`
	IsRecurrent bool    `json:"is_recurrent"`
	StartMonth  string  `json:"start_month" validate:"required"`
	EndMonth    string  `json:"end_month,omitempty"`
	Budget      float64 `json:"budget" validate:"min=0"`
}

type CreateExpenseRequest struct {
	CategoryID  string     `json:"category_id" validate:"required"`
	Amount      float64    `json:"amount" validate:"required,gt=0"`
	Description string     `json:"description" validate:"max=255"`
	SpentAt     time.Time  `json:"spent_at" validate:"required"`
	IsPaid      bool       `json:"is_paid"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
}

type UpdateExpenseRequest struct {
	ID          string     `json:"-"`
	CategoryID  string     `json:"category_id" validate:"required"`
	Amount      float64    `json:"amount" validate:"required,gt=0"`
	Description string     `json:"description" validate:"max=255"`
	SpentAt     time.Time  `json:"spent_at" validate:"required"`
	IsPaid      bool       `json:"is_paid"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
}

type ExpenseResponse struct {
	ID          string     `json:"id"`
	CategoryID  string     `json:"category_id"`
	Amount      float64    `json:"amount"`
	Description string     `json:"description"`
	SpentAt     time.Time  `json:"spent_at"`
	IsPaid      bool       `json:"is_paid"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
}
