package usecase

import (
	"context"

	"github.com/madalinpopa/gocost-web/internal/domain"
	"github.com/madalinpopa/gocost-web/internal/domain/expense"
	"github.com/madalinpopa/gocost-web/internal/domain/identity"
	"github.com/madalinpopa/gocost-web/internal/domain/income"
	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/stretchr/testify/mock"
)

// MockUnitOfWork is a test double for the UnitOfWork interface.
type MockUnitOfWork struct {
	mock.Mock
	UserRepo     *MockUserRepository
	IncomeRepo   *MockIncomeRepository
	ExpenseRepo  *MockExpenseRepository
	TrackingRepo *MockGroupRepository
}

func (m *MockUnitOfWork) UserRepository() identity.UserRepository {
	return m.UserRepo
}

func (m *MockUnitOfWork) IncomeRepository() income.IncomeRepository {
	return m.IncomeRepo
}

func (m *MockUnitOfWork) ExpenseRepository() expense.ExpenseRepository {
	return m.ExpenseRepo
}

func (m *MockUnitOfWork) TrackingRepository() tracking.GroupRepository {
	return m.TrackingRepo
}

func (m *MockUnitOfWork) Begin(ctx context.Context) (domain.UnitOfWork, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.UnitOfWork), args.Error(1)
}

func (m *MockUnitOfWork) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockUnitOfWork) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

// MockUserRepository is a test double for identity.UserRepository.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Save(ctx context.Context, user identity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id identity.ID) (identity.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(identity.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email identity.EmailVO) (identity.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(identity.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username identity.UsernameVO) (identity.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(identity.User), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email identity.EmailVO) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username identity.UsernameVO) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

// MockIncomeRepository is a test double for income.IncomeRepository.
type MockIncomeRepository struct {
	mock.Mock
}

func (m *MockIncomeRepository) Save(ctx context.Context, income income.Income) error {
	args := m.Called(ctx, income)
	return args.Error(0)
}

func (m *MockIncomeRepository) FindByID(ctx context.Context, id income.ID) (income.Income, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(income.Income), args.Error(1)
}

func (m *MockIncomeRepository) FindByUserID(ctx context.Context, userID income.ID) ([]income.Income, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]income.Income), args.Error(1)
}

func (m *MockIncomeRepository) Delete(ctx context.Context, id income.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockGroupRepository is a test double for tracking.GroupRepository.
type MockGroupRepository struct {
	mock.Mock
}

func (m *MockGroupRepository) Save(ctx context.Context, group tracking.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockGroupRepository) FindByID(ctx context.Context, id tracking.ID) (tracking.Group, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(tracking.Group), args.Error(1)
}

func (m *MockGroupRepository) FindByUserID(ctx context.Context, userID tracking.ID) ([]tracking.Group, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]tracking.Group), args.Error(1)
}

func (m *MockGroupRepository) FindGroupByCategoryID(ctx context.Context, categoryID tracking.ID) (tracking.Group, error) {
	args := m.Called(ctx, categoryID)
	return args.Get(0).(tracking.Group), args.Error(1)
}

func (m *MockGroupRepository) Delete(ctx context.Context, id tracking.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGroupRepository) DeleteCategory(ctx context.Context, id tracking.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockExpenseRepository is a test double for expense.ExpenseRepository.
type MockExpenseRepository struct {
	mock.Mock
}

func (m *MockExpenseRepository) Save(ctx context.Context, expense expense.Expense) error {
	args := m.Called(ctx, expense)
	return args.Error(0)
}

func (m *MockExpenseRepository) FindByID(ctx context.Context, id expense.ID) (expense.Expense, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(expense.Expense), args.Error(1)
}

func (m *MockExpenseRepository) FindByUserID(ctx context.Context, userID expense.ID) ([]expense.Expense, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]expense.Expense), args.Error(1)
}

func (m *MockExpenseRepository) FindByUserIDAndMonth(ctx context.Context, userID expense.ID, month string) ([]expense.Expense, error) {
	args := m.Called(ctx, userID, month)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]expense.Expense), args.Error(1)
}

func (m *MockExpenseRepository) Delete(ctx context.Context, id expense.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockExpenseRepository) Total(ctx context.Context, userID expense.ID, month string) (money.Money, error) {
	args := m.Called(ctx, userID, month)
	return args.Get(0).(money.Money), args.Error(1)
}