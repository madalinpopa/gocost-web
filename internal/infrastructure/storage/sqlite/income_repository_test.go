package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/madalinpopa/gocost-web/internal/domain/income"
	"github.com/madalinpopa/gocost-web/internal/infrastructure/storage/sqlite"
	"github.com/madalinpopa/gocost-web/internal/shared/identifier"
	"github.com/madalinpopa/gocost-web/internal/shared/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteIncomeRepository(t *testing.T) {
	repo := sqlite.NewSQLiteIncomeRepository(testDB)
	userRepo := sqlite.NewSQLiteUserRepository(testDB)
	ctx := context.Background()

	t.Run("Save_Success", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))

		inc := createRandomIncome(t, user.ID)
		err := repo.Save(ctx, *inc)
		assert.NoError(t, err)
	})

	t.Run("FindByID_Success", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))

		inc := createRandomIncome(t, user.ID)
		require.NoError(t, repo.Save(ctx, *inc))

		foundIncome, err := repo.FindByID(ctx, inc.ID)
		assert.NoError(t, err)
		assert.Equal(t, inc.ID, foundIncome.ID)
		assert.Equal(t, inc.UserID, foundIncome.UserID)
		assert.Equal(t, inc.Amount, foundIncome.Amount)
		assert.Equal(t, inc.Source, foundIncome.Source)
		assert.WithinDuration(t, inc.ReceivedAt, foundIncome.ReceivedAt, time.Second)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		randomID, _ := identifier.NewID()
		_, err := repo.FindByID(ctx, randomID)
		assert.ErrorIs(t, err, income.ErrIncomeNotFound)
	})

	t.Run("Save_Update", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))

		inc := createRandomIncome(t, user.ID)
		require.NoError(t, repo.Save(ctx, *inc))

		newAmount, _ := money.New(5000)
		newSource, _ := income.NewSourceVO("Bonus")

		// Create a new income instance with updated values but same ID
		updatedInc, err := income.NewIncome(inc.ID, inc.UserID, newAmount, newSource, inc.ReceivedAt)
		require.NoError(t, err)

		err = repo.Save(ctx, *updatedInc)
		assert.NoError(t, err)

		foundIncome, err := repo.FindByID(ctx, inc.ID)
		assert.NoError(t, err)
		assert.Equal(t, newAmount, foundIncome.Amount)
		assert.Equal(t, newSource, foundIncome.Source)
	})

	t.Run("FindByUserID_Success", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))

		inc1 := createRandomIncome(t, user.ID)
		require.NoError(t, repo.Save(ctx, *inc1))

		// Ensure strictly newer time for second income to test sorting
		time.Sleep(time.Millisecond * 10) 
		inc2 := createRandomIncome(t, user.ID)
		require.NoError(t, repo.Save(ctx, *inc2))

		// Create another user and income to ensure we don't fetch them
		otherUser := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *otherUser))
		otherInc := createRandomIncome(t, otherUser.ID)
		require.NoError(t, repo.Save(ctx, *otherInc))

		incomes, err := repo.FindByUserID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Len(t, incomes, 2)
		
		// Expect DESC order by received_at (inc2 is newer)
		assert.Equal(t, inc2.ID, incomes[0].ID)
		assert.Equal(t, inc1.ID, incomes[1].ID)
	})

	t.Run("FindByUserID_Empty", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))

		incomes, err := repo.FindByUserID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Empty(t, incomes)
	})

	t.Run("Delete_Success", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))

		inc := createRandomIncome(t, user.ID)
		require.NoError(t, repo.Save(ctx, *inc))

		err := repo.Delete(ctx, inc.ID)
		assert.NoError(t, err)

		_, err = repo.FindByID(ctx, inc.ID)
		assert.ErrorIs(t, err, income.ErrIncomeNotFound)
	})

	t.Run("Delete_NotFound", func(t *testing.T) {
		randomID, _ := identifier.NewID()
		err := repo.Delete(ctx, randomID)
		assert.ErrorIs(t, err, income.ErrIncomeNotFound)
	})
}
