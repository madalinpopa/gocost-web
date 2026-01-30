package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/madalinpopa/gocost-web/internal/domain/income"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestIncomeUseCase(repo *MockIncomeRepository, userRepo *MockUserRepository) IncomeUseCaseImpl {
	if repo == nil {
		repo = &MockIncomeRepository{}
	}
	if userRepo == nil {
		userRepo = &MockUserRepository{}
	}
	return NewIncomeUseCase(
		&MockUnitOfWork{IncomeRepo: repo, UserRepo: userRepo},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func newTestIncome(t *testing.T, userID identifier.ID) *income.Income {
	t.Helper()

	id, err := identifier.NewID()
	require.NoError(t, err)

	amount, err := money.NewFromFloat(100.50, "USD")
	require.NoError(t, err)

	source, err := income.NewSourceVO("Salary")
	require.NoError(t, err)

	inc, err := income.NewIncome(id, userID, amount, source, time.Now())
	require.NoError(t, err)

	return inc
}

func TestIncomeUseCase_Create(t *testing.T) {
	validUserID, _ := identifier.NewID()
	validReq := &CreateIncomeRequest{
		UserID:     validUserID.String(),
		Currency:   "USD",
		Amount:     100.50,
		Source:     "Salary",
		ReceivedAt: time.Now(),
	}

	t.Run("returns error for nil request", func(t *testing.T) {
		usecase := newTestIncomeUseCase(nil, nil)
		resp, err := usecase.Create(context.Background(), nil)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "request cannot be nil")
	})

	t.Run("returns error for invalid user ID", func(t *testing.T) {
		usecase := newTestIncomeUseCase(nil, nil)
		req := *validReq
		req.UserID = "invalid-id"
		resp, err := usecase.Create(context.Background(), &req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identifier.ErrInvalidID)
	})

	t.Run("returns error for invalid amount", func(t *testing.T) {
		usecase := newTestIncomeUseCase(nil, nil)
		req := &CreateIncomeRequest{
			UserID:     validUserID.String(),
			Currency:   "USD",
			Amount:     -10.0,
			Source:     "Salary",
			ReceivedAt: time.Now(),
		}
		resp, err := usecase.Create(context.Background(), req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, income.ErrInvalidAmount)
	})

	t.Run("saves income with empty source and returns response", func(t *testing.T) {
		repo := &MockIncomeRepository{}
		repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		usecase := newTestIncomeUseCase(repo, nil)

		req := &CreateIncomeRequest{
			UserID:     validUserID.String(),
			Currency:   "USD",
			Amount:     100.0,
			Source:     "",
			ReceivedAt: time.Now(),
		}
		resp, err := usecase.Create(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "", resp.Source)
	})

	t.Run("returns error when save fails", func(t *testing.T) {
		expectedErr := errors.New("save failed")
		repo := &MockIncomeRepository{}
		repo.On("Save", mock.Anything, mock.Anything).Return(expectedErr)

		usecase := newTestIncomeUseCase(repo, nil)

		resp, err := usecase.Create(context.Background(), validReq)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("saves income and returns response", func(t *testing.T) {
		var savedIncome income.Income
		repo := &MockIncomeRepository{}
		repo.On("Save", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			savedIncome = args.Get(1).(income.Income)
		})

		usecase := newTestIncomeUseCase(repo, nil)

		resp, err := usecase.Create(context.Background(), validReq)

		require.NoError(t, err)
		require.NotNil(t, resp)
		expectedAmount, err := money.NewFromFloat(validReq.Amount, validReq.Currency)
		require.NoError(t, err)
		assert.Equal(t, expectedAmount.Cents(), resp.AmountCents)
		assert.Equal(t, validReq.Currency, resp.Currency)
		assert.Equal(t, validReq.Source, resp.Source)
		assert.Equal(t, validReq.ReceivedAt, resp.ReceivedAt)
		assert.NotEmpty(t, resp.ID)

		assert.Equal(t, expectedAmount.Cents(), savedIncome.Amount.Cents())
		assert.Equal(t, validReq.Source, savedIncome.Source.Value())
		assert.Equal(t, validReq.ReceivedAt, savedIncome.ReceivedAt)
		assert.Equal(t, validUserID, savedIncome.UserID)
	})
}

func TestIncomeUseCase_Update(t *testing.T) {
	validUserID, _ := identifier.NewID()
	existingIncome := newTestIncome(t, validUserID)
	validReq := &UpdateIncomeRequest{
		ID:         existingIncome.ID.String(),
		UserID:     validUserID.String(),
		Currency:   "USD",
		Amount:     200.0,
		Source:     "Bonus",
		ReceivedAt: time.Now(),
	}

	t.Run("returns error for nil request", func(t *testing.T) {
		usecase := newTestIncomeUseCase(nil, nil)
		resp, err := usecase.Update(context.Background(), nil)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "request cannot be nil")
	})

	t.Run("returns error for invalid income ID", func(t *testing.T) {
		usecase := newTestIncomeUseCase(nil, nil)
		req := &UpdateIncomeRequest{ID: "invalid", UserID: validUserID.String(), Currency: "USD"}
		resp, err := usecase.Update(context.Background(), req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identifier.ErrInvalidID)
	})

	t.Run("returns error when income not found", func(t *testing.T) {
		expectedErr := errors.New("not found")
		repo := &MockIncomeRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(income.Income{}, expectedErr)

		usecase := newTestIncomeUseCase(repo, nil)
		resp, err := usecase.Update(context.Background(), validReq)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("returns error when user unauthorized", func(t *testing.T) {
		otherUserID, _ := identifier.NewID()
		otherUserIncome := newTestIncome(t, otherUserID)
		repo := &MockIncomeRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*otherUserIncome, nil)

		usecase := newTestIncomeUseCase(repo, nil)

		req := &UpdateIncomeRequest{
			ID: otherUserIncome.ID.String(), UserID: validUserID.String(), Currency: "USD", Amount: 100, Source: "Test", ReceivedAt: time.Now(),
		}
		resp, err := usecase.Update(context.Background(), req) // validUserID != otherUserID

		assert.Nil(t, resp)
		assert.EqualError(t, err, "unauthorized")
	})

	t.Run("updates income and returns response", func(t *testing.T) {
		var savedIncome income.Income
		repo := &MockIncomeRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*existingIncome, nil)
		repo.On("Save", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			savedIncome = args.Get(1).(income.Income)
		})

		usecase := newTestIncomeUseCase(repo, nil)

		resp, err := usecase.Update(context.Background(), validReq)

		require.NoError(t, err)
		require.NotNil(t, resp)
		expectedAmount, err := money.NewFromFloat(validReq.Amount, validReq.Currency)
		require.NoError(t, err)
		assert.Equal(t, expectedAmount.Cents(), resp.AmountCents)
		assert.Equal(t, validReq.Currency, resp.Currency)
		assert.Equal(t, validReq.Source, resp.Source)
		assert.Equal(t, existingIncome.ID.String(), resp.ID)

		assert.Equal(t, expectedAmount.Cents(), savedIncome.Amount.Cents())
		assert.Equal(t, validReq.Source, savedIncome.Source.Value())
	})
}

func TestIncomeUseCase_Delete(t *testing.T) {
	validUserID, _ := identifier.NewID()
	existingIncome := newTestIncome(t, validUserID)

	t.Run("returns error for invalid income ID", func(t *testing.T) {
		usecase := newTestIncomeUseCase(nil, nil)
		err := usecase.Delete(context.Background(), validUserID.String(), "invalid")
		assert.ErrorIs(t, err, identifier.ErrInvalidID)
	})

	t.Run("returns error when income not found", func(t *testing.T) {
		expectedErr := errors.New("not found")
		repo := &MockIncomeRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(income.Income{}, expectedErr)

		usecase := newTestIncomeUseCase(repo, nil)
		err := usecase.Delete(context.Background(), validUserID.String(), existingIncome.ID.String())
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("returns error when unauthorized", func(t *testing.T) {
		otherUserID, _ := identifier.NewID()
		otherUserIncome := newTestIncome(t, otherUserID)
		repo := &MockIncomeRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*otherUserIncome, nil)

		usecase := newTestIncomeUseCase(repo, nil)
		err := usecase.Delete(context.Background(), validUserID.String(), otherUserIncome.ID.String())
		assert.EqualError(t, err, "unauthorized")
	})

	t.Run("deletes income successfully", func(t *testing.T) {
		var deletedID identifier.ID
		repo := &MockIncomeRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*existingIncome, nil)
		repo.On("Delete", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			deletedID = args.Get(1).(identifier.ID)
		})

		usecase := newTestIncomeUseCase(repo, nil)

		err := usecase.Delete(context.Background(), validUserID.String(), existingIncome.ID.String())
		require.NoError(t, err)
		assert.Equal(t, existingIncome.ID, deletedID)
	})
}

func TestIncomeUseCase_Get(t *testing.T) {
	validUserID, _ := identifier.NewID()
	existingIncome := newTestIncome(t, validUserID)

	t.Run("returns income successfully", func(t *testing.T) {
		repo := &MockIncomeRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*existingIncome, nil)

		usecase := newTestIncomeUseCase(repo, nil)

		resp, err := usecase.Get(context.Background(), validUserID.String(), existingIncome.ID.String())
		require.NoError(t, err)
		assert.Equal(t, existingIncome.ID.String(), resp.ID)
		assert.Equal(t, existingIncome.Amount.Cents(), resp.AmountCents)
		assert.Equal(t, existingIncome.Amount.Currency(), resp.Currency)
	})

	t.Run("returns unauthorized for different user", func(t *testing.T) {
		otherUserID, _ := identifier.NewID()
		repo := &MockIncomeRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*existingIncome, nil)

		usecase := newTestIncomeUseCase(repo, nil)

		resp, err := usecase.Get(context.Background(), otherUserID.String(), existingIncome.ID.String())
		assert.Nil(t, resp)
		assert.EqualError(t, err, "unauthorized")
	})
}

func TestIncomeUseCase_List(t *testing.T) {
	validUserID, _ := identifier.NewID()
	inc1 := newTestIncome(t, validUserID)
	inc2 := newTestIncome(t, validUserID)

	t.Run("returns list of incomes", func(t *testing.T) {
		repo := &MockIncomeRepository{}
		repo.On("FindByUserID", mock.Anything, mock.Anything).Return([]income.Income{*inc1, *inc2}, nil)

		usecase := newTestIncomeUseCase(repo, nil)

		resps, err := usecase.List(context.Background(), validUserID.String())
		require.NoError(t, err)
		assert.Len(t, resps, 2)
		assert.Equal(t, inc1.ID.String(), resps[0].ID)
		assert.Equal(t, inc2.ID.String(), resps[1].ID)
	})

	t.Run("returns empty list if none found", func(t *testing.T) {
		repo := &MockIncomeRepository{}
		repo.On("FindByUserID", mock.Anything, mock.Anything).Return([]income.Income{}, nil)

		usecase := newTestIncomeUseCase(repo, nil)

		resps, err := usecase.List(context.Background(), validUserID.String())

		require.NoError(t, err)

		assert.Empty(t, resps)

	})

}

func TestIncomeUseCase_Total(t *testing.T) {

	validUserID, _ := identifier.NewID()

	// Create incomes for different months

	inc1 := newTestIncome(t, validUserID)

	currentMonth, _ := time.Parse("2006-01", "2023-10")

	inc1.ReceivedAt = currentMonth.AddDate(0, 0, 1) // 2023-10-02

	inc2 := newTestIncome(t, validUserID)

	inc2.ReceivedAt = currentMonth.AddDate(0, 0, 15) // 2023-10-16

	inc3 := newTestIncome(t, validUserID)

	inc3.ReceivedAt = currentMonth.AddDate(0, 1, 1) // 2023-11-02 (Next month)

	t.Run("returns total amount for specific month", func(t *testing.T) {

		repo := &MockIncomeRepository{}

		expectedTotalMoney, err := money.NewFromFloat(inc1.Amount.Amount()+inc2.Amount.Amount(), "USD")
		require.NoError(t, err)
		repo.On("TotalByUserIDAndMonth", mock.Anything, validUserID, "2023-10").Return(expectedTotalMoney, nil)

		usecase := newTestIncomeUseCase(repo, nil)

		total, err := usecase.Total(context.Background(), validUserID.String(), "2023-10")

		require.NoError(t, err)

		assert.Equal(t, expectedTotalMoney.Amount(), total)

	})

	t.Run("returns zero if no incomes in month", func(t *testing.T) {

		repo := &MockIncomeRepository{}

		zeroMoney, err := money.New(0, "USD")
		require.NoError(t, err)
		repo.On("TotalByUserIDAndMonth", mock.Anything, validUserID, "2023-12").Return(zeroMoney, nil)

		usecase := newTestIncomeUseCase(repo, nil)

		total, err := usecase.Total(context.Background(), validUserID.String(), "2023-12")

		require.NoError(t, err)

		assert.Equal(t, 0.0, total)

	})

	t.Run("returns error for invalid date format", func(t *testing.T) {

		expectedErr := errors.New("invalid month")
		repo := &MockIncomeRepository{}
		repo.On("TotalByUserIDAndMonth", mock.Anything, validUserID, "invalid-date").Return(money.Money{}, expectedErr)

		usecase := newTestIncomeUseCase(repo, nil)

		_, err := usecase.Total(context.Background(), validUserID.String(), "invalid-date")
		assert.ErrorIs(t, err, expectedErr)

	})

}
