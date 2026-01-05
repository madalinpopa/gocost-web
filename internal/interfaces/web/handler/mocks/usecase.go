package mocks

import (
	"context"

	"github.com/madalinpopa/gocost-web/internal/usecase"
	"github.com/stretchr/testify/mock"
)

type MockAuthUseCase struct {
	mock.Mock
}

func (m *MockAuthUseCase) Register(ctx context.Context, req *usecase.RegisterUserRequest) (*usecase.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.UserResponse), args.Error(1)
}

func (m *MockAuthUseCase) Login(ctx context.Context, req *usecase.LoginRequest) (*usecase.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.LoginResponse), args.Error(1)
}

type MockIncomeUseCase struct {
	mock.Mock
}

func (m *MockIncomeUseCase) Create(ctx context.Context, userID string, req *usecase.CreateIncomeRequest) (*usecase.IncomeResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.IncomeResponse), args.Error(1)
}

func (m *MockIncomeUseCase) Update(ctx context.Context, userID string, req *usecase.UpdateIncomeRequest) (*usecase.IncomeResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.IncomeResponse), args.Error(1)
}

func (m *MockIncomeUseCase) Delete(ctx context.Context, userID string, id string) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

func (m *MockIncomeUseCase) Get(ctx context.Context, userID string, id string) (*usecase.IncomeResponse, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.IncomeResponse), args.Error(1)
}

func (m *MockIncomeUseCase) List(ctx context.Context, userID string) ([]*usecase.IncomeResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*usecase.IncomeResponse), args.Error(1)
}

func (m *MockIncomeUseCase) ListByMonth(ctx context.Context, userID string, month string) ([]*usecase.IncomeResponse, error) {
	args := m.Called(ctx, userID, month)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*usecase.IncomeResponse), args.Error(1)
}

func (m *MockIncomeUseCase) Total(ctx context.Context, userID string, month string) (float64, error) {
	args := m.Called(ctx, userID, month)
	return args.Get(0).(float64), args.Error(1)
}

type MockGroupUseCase struct {
	mock.Mock
}

func (m *MockGroupUseCase) Create(ctx context.Context, userID string, req *usecase.CreateGroupRequest) (*usecase.GroupResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.GroupResponse), args.Error(1)
}

func (m *MockGroupUseCase) Update(ctx context.Context, userID string, req *usecase.UpdateGroupRequest) (*usecase.GroupResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.GroupResponse), args.Error(1)
}

func (m *MockGroupUseCase) Delete(ctx context.Context, userID string, id string) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

func (m *MockGroupUseCase) Get(ctx context.Context, userID string, id string) (*usecase.GroupResponse, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.GroupResponse), args.Error(1)
}

func (m *MockGroupUseCase) List(ctx context.Context, userID string) ([]*usecase.GroupResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*usecase.GroupResponse), args.Error(1)
}

type MockCategoryUseCase struct {
	mock.Mock
}

func (m *MockCategoryUseCase) Create(ctx context.Context, groupID string, req *usecase.CreateCategoryRequest) (*usecase.CategoryResponse, error) {
	args := m.Called(ctx, groupID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.CategoryResponse), args.Error(1)
}

func (m *MockCategoryUseCase) Update(ctx context.Context, groupID string, req *usecase.UpdateCategoryRequest) (*usecase.CategoryResponse, error) {
	args := m.Called(ctx, groupID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.CategoryResponse), args.Error(1)
}

func (m *MockCategoryUseCase) Delete(ctx context.Context, groupID string, id string) error {
	args := m.Called(ctx, groupID, id)
	return args.Error(0)
}

func (m *MockCategoryUseCase) Get(ctx context.Context, groupID string, id string) (*usecase.CategoryResponse, error) {
	args := m.Called(ctx, groupID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.CategoryResponse), args.Error(1)
}

func (m *MockCategoryUseCase) List(ctx context.Context, groupID string) ([]usecase.CategoryResponse, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]usecase.CategoryResponse), args.Error(1)
}

type MockExpenseUseCase struct {
	mock.Mock
}

func (m *MockExpenseUseCase) Create(ctx context.Context, userID string, req *usecase.CreateExpenseRequest) (*usecase.ExpenseResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.ExpenseResponse), args.Error(1)
}

func (m *MockExpenseUseCase) Update(ctx context.Context, userID string, req *usecase.UpdateExpenseRequest) (*usecase.ExpenseResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.ExpenseResponse), args.Error(1)
}

func (m *MockExpenseUseCase) Delete(ctx context.Context, userID string, id string) error {
	args := m.Called(ctx, userID, id)
	return args.Error(0)
}

func (m *MockExpenseUseCase) Get(ctx context.Context, userID string, id string) (*usecase.ExpenseResponse, error) {
	args := m.Called(ctx, userID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.ExpenseResponse), args.Error(1)
}

func (m *MockExpenseUseCase) List(ctx context.Context, userID string) ([]*usecase.ExpenseResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*usecase.ExpenseResponse), args.Error(1)
}

func (m *MockExpenseUseCase) ListByMonth(ctx context.Context, userID string, month string) ([]*usecase.ExpenseResponse, error) {
	args := m.Called(ctx, userID, month)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*usecase.ExpenseResponse), args.Error(1)
}

func (m *MockExpenseUseCase) Total(ctx context.Context, userID string, month string) (float64, error) {
	args := m.Called(ctx, userID, month)
	return args.Get(0).(float64), args.Error(1)
}
