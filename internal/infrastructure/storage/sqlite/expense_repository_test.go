package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/madalinpopa/gocost-web/internal/domain/expense"
	"github.com/madalinpopa/gocost-web/internal/infrastructure/storage/sqlite"
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteExpenseRepository(t *testing.T) {
	repo := sqlite.NewSQLiteExpenseRepository(testDB)
	userRepo := sqlite.NewSQLiteUserRepository(testDB)
	ctx := context.Background()

	t.Run("Save_Success", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))
		group := createRandomGroup(t, user.ID)
		category := createRandomCategory(t, group.ID)

		exp := createRandomExpense(t, category.ID)
		err := repo.Save(ctx, *exp)
		assert.NoError(t, err)
	})

	t.Run("FindByID_Success", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))
		group := createRandomGroup(t, user.ID)
		category := createRandomCategory(t, group.ID)

		exp := createRandomExpense(t, category.ID)
		require.NoError(t, repo.Save(ctx, *exp))

		foundExpense, err := repo.FindByID(ctx, exp.ID)
		assert.NoError(t, err)
		assert.Equal(t, exp.ID, foundExpense.ID)
		assert.Equal(t, exp.CategoryID, foundExpense.CategoryID)
		assert.Equal(t, exp.Amount, foundExpense.Amount)
		assert.Equal(t, exp.Description, foundExpense.Description)
		assert.WithinDuration(t, exp.SpentAt, foundExpense.SpentAt, time.Second)
		assert.False(t, foundExpense.Payment.IsPaid())
		assert.Nil(t, foundExpense.Payment.PaidAt())
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		randomID, _ := identifier.NewID()
		_, err := repo.FindByID(ctx, randomID)
		assert.ErrorIs(t, err, expense.ErrExpenseNotFound)
	})

	t.Run("FindByUserID_Success", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))
		group := createRandomGroup(t, user.ID)
		category := createRandomCategory(t, group.ID)

		exp1 := createRandomExpense(t, category.ID)
		require.NoError(t, repo.Save(ctx, *exp1))

		time.Sleep(10 * time.Millisecond)
		exp2 := createRandomExpense(t, category.ID)
		require.NoError(t, repo.Save(ctx, *exp2))

		// Another user
		otherUser := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *otherUser))
		otherGroup := createRandomGroup(t, otherUser.ID)
		otherCategory := createRandomCategory(t, otherGroup.ID)
		otherExp := createRandomExpense(t, otherCategory.ID)
		require.NoError(t, repo.Save(ctx, *otherExp))

		expenses, err := repo.FindByUserID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Len(t, expenses, 2)
		assert.Equal(t, exp2.ID, expenses[0].ID)
		assert.Equal(t, exp1.ID, expenses[1].ID)
	})

	t.Run("Delete_Success", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))
		group := createRandomGroup(t, user.ID)
		category := createRandomCategory(t, group.ID)

		exp := createRandomExpense(t, category.ID)
		require.NoError(t, repo.Save(ctx, *exp))

		err := repo.Delete(ctx, exp.ID)
		assert.NoError(t, err)

		_, err = repo.FindByID(ctx, exp.ID)
		assert.ErrorIs(t, err, expense.ErrExpenseNotFound)
	})

	t.Run("Save_Update", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))
		group := createRandomGroup(t, user.ID)
		category := createRandomCategory(t, group.ID)

		exp := createRandomExpense(t, category.ID)
		require.NoError(t, repo.Save(ctx, *exp))

		newAmount, _ := money.New(2000, "USD")
		newDesc, _ := expense.NewExpenseDescriptionVO("Dinner")
		paidAt := time.Now()
		payment, err := expense.NewPaidStatus(paidAt)
		require.NoError(t, err)

		updatedExp, err := expense.NewExpense(exp.ID, exp.CategoryID, newAmount, newDesc, exp.SpentAt, payment)
		require.NoError(t, err)

		err = repo.Save(ctx, *updatedExp)
		assert.NoError(t, err)

		foundExpense, err := repo.FindByID(ctx, exp.ID)
		assert.NoError(t, err)
		assert.Equal(t, newAmount, foundExpense.Amount)
		assert.Equal(t, newDesc, foundExpense.Description)
		assert.True(t, foundExpense.Payment.IsPaid())
		require.NotNil(t, foundExpense.Payment.PaidAt())
		assert.WithinDuration(t, paidAt, *foundExpense.Payment.PaidAt(), time.Second)
	})

	t.Run("FindByUserIDAndMonth_Success", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))
		group := createRandomGroup(t, user.ID)
		category := createRandomCategory(t, group.ID)

		// Expense in Oct 2023
		exp1 := createRandomExpense(t, category.ID)
		exp1.SpentAt = time.Date(2023, 10, 5, 0, 0, 0, 0, time.UTC)
		require.NoError(t, repo.Save(ctx, *exp1))

		// Expense in Nov 2023
		exp2 := createRandomExpense(t, category.ID)
		exp2.SpentAt = time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC)
		require.NoError(t, repo.Save(ctx, *exp2))

		expenses, err := repo.FindByUserIDAndMonth(ctx, user.ID, "2023-10")
		assert.NoError(t, err)
		assert.Len(t, expenses, 1)
		assert.Equal(t, exp1.ID, expenses[0].ID)

		expensesNov, err := repo.FindByUserIDAndMonth(ctx, user.ID, "2023-11")
		assert.NoError(t, err)
		assert.Len(t, expensesNov, 1)
		assert.Equal(t, exp2.ID, expensesNov[0].ID)
	})

	t.Run("Total_Success", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))
		group := createRandomGroup(t, user.ID)
		category := createRandomCategory(t, group.ID)

		// Expense in Oct 2023
		exp1 := createRandomExpense(t, category.ID)
		exp1.SpentAt = time.Date(2023, 10, 5, 0, 0, 0, 0, time.UTC)
		require.NoError(t, repo.Save(ctx, *exp1))

		// Another expense in Oct 2023
		exp2 := createRandomExpense(t, category.ID)
		exp2.SpentAt = time.Date(2023, 10, 15, 0, 0, 0, 0, time.UTC)
		require.NoError(t, repo.Save(ctx, *exp2))

		// Expense in Nov 2023
		exp3 := createRandomExpense(t, category.ID)
		exp3.SpentAt = time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC)
		require.NoError(t, repo.Save(ctx, *exp3))

		total, err := repo.Total(ctx, user.ID, "2023-10")
		assert.NoError(t, err)
		
		expectedTotal, _ := exp1.Amount.Add(exp2.Amount)
		isEqual, err := expectedTotal.Equals(total)
	assert.NoError(t, err)
	assert.True(t, isEqual)

		totalNov, err := repo.Total(ctx, user.ID, "2023-11")
		assert.NoError(t, err)
		isEqual, err = exp3.Amount.Equals(totalNov)
	assert.NoError(t, err)
	assert.True(t, isEqual)

		totalDec, err := repo.Total(ctx, user.ID, "2023-12")
		assert.NoError(t, err)
		isZero, err := totalDec.IsZero()
	assert.NoError(t, err)
	assert.True(t, isZero)
	})
}