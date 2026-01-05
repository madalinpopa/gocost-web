package usecase

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/madalinpopa/gocost-web/internal/domain/income"
	"github.com/madalinpopa/gocost-web/internal/shared/identifier"
	"github.com/madalinpopa/gocost-web/internal/shared/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestIncomeUseCase(repo *MockIncomeRepository) IncomeUseCaseImpl {
	if repo == nil {
		repo = &MockIncomeRepository{}
	}
	return NewIncomeUseCase(
		&MockUnitOfWork{IncomeRepo: repo},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
}

func newTestIncome(t *testing.T, userID identifier.ID) *income.Income {
	t.Helper()

	id, err := identifier.NewID()
	require.NoError(t, err)

	amount, err := money.NewFromFloat(100.50)
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
		Amount:     100.50,
		Source:     "Salary",
		ReceivedAt: time.Now(),
	}

	t.Run("returns error for nil request", func(t *testing.T) {
		usecase := newTestIncomeUseCase(nil)
		resp, err := usecase.Create(context.Background(), validUserID.String(), nil)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "request cannot be nil")
	})

	t.Run("returns error for invalid user ID", func(t *testing.T) {
		usecase := newTestIncomeUseCase(nil)
		resp, err := usecase.Create(context.Background(), "invalid-id", validReq)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identifier.ErrInvalidID)
	})

	t.Run("returns error for invalid amount", func(t *testing.T) {
		usecase := newTestIncomeUseCase(nil)
		req := &CreateIncomeRequest{
			Amount:     -10.0,
			Source:     "Salary",
			ReceivedAt: time.Now(),
		}
		resp, err := usecase.Create(context.Background(), validUserID.String(), req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, money.ErrNegativeAmount)
	})

	t.Run("saves income with empty source and returns response", func(t *testing.T) {
		repo := &MockIncomeRepository{}
		repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		usecase := newTestIncomeUseCase(repo)

		req := &CreateIncomeRequest{
			Amount:     100.0,
			Source:     "",
			ReceivedAt: time.Now(),
		}
		resp, err := usecase.Create(context.Background(), validUserID.String(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "", resp.Source)
	})

	t.Run("returns error when save fails", func(t *testing.T) {
		expectedErr := errors.New("save failed")
		repo := &MockIncomeRepository{}
		repo.On("Save", mock.Anything, mock.Anything).Return(expectedErr)

		usecase := newTestIncomeUseCase(repo)

		resp, err := usecase.Create(context.Background(), validUserID.String(), validReq)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("saves income and returns response", func(t *testing.T) {
		var savedIncome income.Income
		repo := &MockIncomeRepository{}
		repo.On("Save", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			savedIncome = args.Get(1).(income.Income)
		})

		usecase := newTestIncomeUseCase(repo)

		resp, err := usecase.Create(context.Background(), validUserID.String(), validReq)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, validReq.Amount, resp.Amount)
		assert.Equal(t, validReq.Source, resp.Source)
		assert.Equal(t, validReq.ReceivedAt, resp.ReceivedAt)
		assert.NotEmpty(t, resp.ID)

		assert.Equal(t, validReq.Amount, savedIncome.Amount.Amount())
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
		Amount:     200.0,
		Source:     "Bonus",
		ReceivedAt: time.Now(),
	}

	t.Run("returns error for nil request", func(t *testing.T) {
		usecase := newTestIncomeUseCase(nil)
		resp, err := usecase.Update(context.Background(), validUserID.String(), nil)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "request cannot be nil")
	})

	t.Run("returns error for invalid income ID", func(t *testing.T) {
		usecase := newTestIncomeUseCase(nil)
		req := &UpdateIncomeRequest{ID: "invalid"}
		resp, err := usecase.Update(context.Background(), validUserID.String(), req)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, identifier.ErrInvalidID)
	})

	t.Run("returns error when income not found", func(t *testing.T) {
		expectedErr := errors.New("not found")
		repo := &MockIncomeRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(income.Income{}, expectedErr)

		usecase := newTestIncomeUseCase(repo)
		resp, err := usecase.Update(context.Background(), validUserID.String(), validReq)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("returns error when user unauthorized", func(t *testing.T) {
		otherUserID, _ := identifier.NewID()
		otherUserIncome := newTestIncome(t, otherUserID)
		repo := &MockIncomeRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*otherUserIncome, nil)

		usecase := newTestIncomeUseCase(repo)

		req := &UpdateIncomeRequest{ID: otherUserIncome.ID.String(), Amount: 100, Source: "Test", ReceivedAt: time.Now()}
		resp, err := usecase.Update(context.Background(), validUserID.String(), req) // validUserID != otherUserID

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

		usecase := newTestIncomeUseCase(repo)

		resp, err := usecase.Update(context.Background(), validUserID.String(), validReq)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, validReq.Amount, resp.Amount)
		assert.Equal(t, validReq.Source, resp.Source)
		assert.Equal(t, existingIncome.ID.String(), resp.ID)

		assert.Equal(t, validReq.Amount, savedIncome.Amount.Amount())
		assert.Equal(t, validReq.Source, savedIncome.Source.Value())
	})
}

func TestIncomeUseCase_Delete(t *testing.T) {
	validUserID, _ := identifier.NewID()
	existingIncome := newTestIncome(t, validUserID)

	t.Run("returns error for invalid income ID", func(t *testing.T) {
		usecase := newTestIncomeUseCase(nil)
		err := usecase.Delete(context.Background(), validUserID.String(), "invalid")
		assert.ErrorIs(t, err, identifier.ErrInvalidID)
	})

	t.Run("returns error when income not found", func(t *testing.T) {
		expectedErr := errors.New("not found")
		repo := &MockIncomeRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(income.Income{}, expectedErr)

		usecase := newTestIncomeUseCase(repo)
		err := usecase.Delete(context.Background(), validUserID.String(), existingIncome.ID.String())
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("returns error when unauthorized", func(t *testing.T) {
		otherUserID, _ := identifier.NewID()
		otherUserIncome := newTestIncome(t, otherUserID)
		repo := &MockIncomeRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*otherUserIncome, nil)

		usecase := newTestIncomeUseCase(repo)
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

		usecase := newTestIncomeUseCase(repo)

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

		usecase := newTestIncomeUseCase(repo)

		resp, err := usecase.Get(context.Background(), validUserID.String(), existingIncome.ID.String())
		require.NoError(t, err)
		assert.Equal(t, existingIncome.ID.String(), resp.ID)
		assert.Equal(t, existingIncome.Amount.Amount(), resp.Amount)
	})

	t.Run("returns unauthorized for different user", func(t *testing.T) {
		otherUserID, _ := identifier.NewID()
		repo := &MockIncomeRepository{}
		repo.On("FindByID", mock.Anything, mock.Anything).Return(*existingIncome, nil)

		usecase := newTestIncomeUseCase(repo)

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

		usecase := newTestIncomeUseCase(repo)

		resps, err := usecase.List(context.Background(), validUserID.String())
		require.NoError(t, err)
		assert.Len(t, resps, 2)
		assert.Equal(t, inc1.ID.String(), resps[0].ID)
		assert.Equal(t, inc2.ID.String(), resps[1].ID)
	})

	t.Run("returns empty list if none found", func(t *testing.T) {
		repo := &MockIncomeRepository{}
		repo.On("FindByUserID", mock.Anything, mock.Anything).Return([]income.Income{}, nil)

		usecase := newTestIncomeUseCase(repo)

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

		// We expect FindByUserID to be called and return all incomes

		repo.On("FindByUserID", mock.Anything, validUserID).Return([]income.Income{*inc1, *inc2, *inc3}, nil)

		usecase := newTestIncomeUseCase(repo)

		total, err := usecase.Total(context.Background(), validUserID.String(), "2023-10")

		require.NoError(t, err)

		expectedTotal := inc1.Amount.Amount() + inc2.Amount.Amount()

		assert.Equal(t, expectedTotal, total)

	})

	t.Run("returns zero if no incomes in month", func(t *testing.T) {

		repo := &MockIncomeRepository{}

		repo.On("FindByUserID", mock.Anything, validUserID).Return([]income.Income{*inc1, *inc2, *inc3}, nil)

		usecase := newTestIncomeUseCase(repo)

		total, err := usecase.Total(context.Background(), validUserID.String(), "2023-12")

		require.NoError(t, err)

		assert.Equal(t, 0.0, total)

	})

	t.Run("returns error for invalid date format", func(t *testing.T) {

		usecase := newTestIncomeUseCase(nil)

		_, err := usecase.Total(context.Background(), validUserID.String(), "invalid-date")

		assert.Error(t, err)

	})

}
