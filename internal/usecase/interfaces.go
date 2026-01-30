package usecase

import "context"

type AuthUseCase interface {
	Register(ctx context.Context, req *RegisterUserRequest) (*UserResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
}

type IncomeUseCase interface {
	Create(ctx context.Context, req *CreateIncomeRequest) (*IncomeResponse, error)
	Update(ctx context.Context, req *UpdateIncomeRequest) (*IncomeResponse, error)
	Delete(ctx context.Context, userID string, id string) error
	Get(ctx context.Context, userID string, id string) (*IncomeResponse, error)
	List(ctx context.Context, userID string) ([]*IncomeResponse, error)
	ListByMonth(ctx context.Context, userID string, month string) ([]*IncomeResponse, error)
	Total(ctx context.Context, userID string, month string) (float64, error)
}

type GroupUseCase interface {
	Create(ctx context.Context, req *CreateGroupRequest) (*GroupResponse, error)
	Update(ctx context.Context, req *UpdateGroupRequest) (*GroupResponse, error)
	Delete(ctx context.Context, userID string, id string) error
	Get(ctx context.Context, userID string, id string) (*GroupResponse, error)
	List(ctx context.Context, userID string) ([]*GroupResponse, error)
}

type CategoryUseCase interface {
	Create(ctx context.Context, req *CreateCategoryRequest) (*CategoryResponse, error)
	Update(ctx context.Context, req *UpdateCategoryRequest) (*CategoryResponse, error)
	Delete(ctx context.Context, userID string, groupID string, id string) error
	Get(ctx context.Context, userID string, groupID string, id string) (*CategoryResponse, error)
	List(ctx context.Context, userID string, groupID string) ([]CategoryResponse, error)
}

type ExpenseUseCase interface {
	Create(ctx context.Context, req *CreateExpenseRequest) (*ExpenseResponse, error)
	Update(ctx context.Context, req *UpdateExpenseRequest) (*ExpenseResponse, error)
	Delete(ctx context.Context, userID string, id string) error
	Get(ctx context.Context, userID string, id string) (*ExpenseResponse, error)
	List(ctx context.Context, userID string) ([]*ExpenseResponse, error)
	ListByMonth(ctx context.Context, userID string, month string) ([]*ExpenseResponse, error)
	Total(ctx context.Context, userID string, month string) (float64, error)
}
