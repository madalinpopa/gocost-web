package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/madalinpopa/gocost-web/internal/domain/tracking"
	"github.com/madalinpopa/gocost-web/internal/infrastructure/storage/sqlite"
	"github.com/madalinpopa/gocost-web/internal/shared/identifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteTrackingRepository(t *testing.T) {
	repo := sqlite.NewSQLiteTrackingRepository(testDB)
	userRepo := sqlite.NewSQLiteUserRepository(testDB)
	ctx := context.Background()

	t.Run("SaveAndFindByID_Success", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))

		group := newGroup(t, user.ID, "Personal")
		startMonth := mustMonth(t, 2024, time.January)
		endMonth := mustMonth(t, 2024, time.March)

		catFood := addCategory(t, group, "Food", false, startMonth, tracking.Month{})
		catRent := addCategory(t, group, "Rent", true, startMonth, endMonth)

		require.NoError(t, repo.Save(ctx, *group))

		foundGroup, err := repo.FindByID(ctx, group.ID)
		assert.NoError(t, err)
		assert.Equal(t, group.ID, foundGroup.ID)
		assert.Equal(t, group.UserID, foundGroup.UserID)
		assert.Equal(t, group.Name, foundGroup.Name)
		assert.Equal(t, group.Description, foundGroup.Description)
		assert.Len(t, foundGroup.Categories, 2)

		expected := map[string]*tracking.Category{
			catFood.Name.Value(): catFood,
			catRent.Name.Value(): catRent,
		}

		for _, category := range foundGroup.Categories {
			expectedCategory := expected[category.Name.Value()]
			require.NotNil(t, expectedCategory)
			assert.Equal(t, expectedCategory.ID, category.ID)
			assert.Equal(t, expectedCategory.GroupID, category.GroupID)
			assert.Equal(t, expectedCategory.Description, category.Description)
			assert.Equal(t, expectedCategory.IsRecurrent, category.IsRecurrent)
			assert.Equal(t, expectedCategory.StartMonth, category.StartMonth)
			assert.Equal(t, expectedCategory.EndMonth, category.EndMonth)
		}
	})

	t.Run("FindByUserID_Success", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))

		groupA := newGroup(t, user.ID, "Personal")
		addCategory(t, groupA, "Food", false, mustMonth(t, 2024, time.January), tracking.Month{})
		require.NoError(t, repo.Save(ctx, *groupA))

		groupB := newGroup(t, user.ID, "Household")
		addCategory(t, groupB, "Rent", true, mustMonth(t, 2024, time.January), mustMonth(t, 2024, time.March))
		require.NoError(t, repo.Save(ctx, *groupB))

		otherUser := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *otherUser))
		otherGroup := newGroup(t, otherUser.ID, "Other")
		addCategory(t, otherGroup, "Other", false, mustMonth(t, 2024, time.January), tracking.Month{})
		require.NoError(t, repo.Save(ctx, *otherGroup))

		groups, err := repo.FindByUserID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Len(t, groups, 2)

		groupByID := make(map[identifier.ID]tracking.Group, len(groups))
		for _, group := range groups {
			groupByID[group.ID] = group
		}
		assert.Contains(t, groupByID, groupA.ID)
		assert.Contains(t, groupByID, groupB.ID)
		assert.NotContains(t, groupByID, otherGroup.ID)

		assert.Len(t, groupByID[groupA.ID].Categories, 1)
		assert.Len(t, groupByID[groupB.ID].Categories, 1)
	})

	t.Run("Save_Update", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))

		group := newGroup(t, user.ID, "Personal")
		startMonth := mustMonth(t, 2024, time.January)
		category := addCategory(t, group, "Food", false, startMonth, tracking.Month{})
		require.NoError(t, repo.Save(ctx, *group))

		updatedGroup := newGroupWithID(t, group.ID, user.ID, "Personal Updated")
		updatedStartMonth := mustMonth(t, 2024, time.February)
		updatedEndMonth := mustMonth(t, 2024, time.March)
		addCategoryWithID(t, updatedGroup, category.ID, "Food Updated", true, updatedStartMonth, updatedEndMonth)
		require.NoError(t, repo.Save(ctx, *updatedGroup))

		foundGroup, err := repo.FindByID(ctx, group.ID)
		assert.NoError(t, err)
		assert.Equal(t, updatedGroup.Name, foundGroup.Name)
		assert.Equal(t, updatedGroup.Description, foundGroup.Description)
		assert.Len(t, foundGroup.Categories, 1)
		assert.Equal(t, "Food Updated", foundGroup.Categories[0].Name.Value())
		assert.True(t, foundGroup.Categories[0].IsRecurrent)
		assert.Equal(t, updatedStartMonth, foundGroup.Categories[0].StartMonth)
		assert.Equal(t, updatedEndMonth, foundGroup.Categories[0].EndMonth)
	})

	t.Run("FindGroupByCategoryID_Success", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))

		group := newGroup(t, user.ID, "Personal")
		category := addCategory(t, group, "Food", false, mustMonth(t, 2024, time.January), tracking.Month{})
		require.NoError(t, repo.Save(ctx, *group))

		foundGroup, err := repo.FindGroupByCategoryID(ctx, category.ID)
		assert.NoError(t, err)
		assert.Equal(t, group.ID, foundGroup.ID)
		assert.Equal(t, group.Name, foundGroup.Name)
		assert.Len(t, foundGroup.Categories, 1)
		assert.Equal(t, category.ID, foundGroup.Categories[0].ID)
	})

	t.Run("FindGroupByCategoryID_NotFound", func(t *testing.T) {
		id, err := identifier.NewID()
		require.NoError(t, err)

		_, err = repo.FindGroupByCategoryID(ctx, id)
		assert.ErrorIs(t, err, tracking.ErrGroupNotFound)
	})

	t.Run("Delete_Success", func(t *testing.T) {
		user := createRandomUser(t)
		require.NoError(t, userRepo.Save(ctx, *user))

		group := newGroup(t, user.ID, "Personal")
		addCategory(t, group, "Food", false, mustMonth(t, 2024, time.January), tracking.Month{})
		require.NoError(t, repo.Save(ctx, *group))

		err := repo.Delete(ctx, group.ID)
		assert.NoError(t, err)

		_, err = repo.FindByID(ctx, group.ID)
		assert.ErrorIs(t, err, tracking.ErrGroupNotFound)
	})
}

func newGroup(t *testing.T, userID identifier.ID, name string) *tracking.Group {
	t.Helper()
	id, err := identifier.NewID()
	require.NoError(t, err)
	return newGroupWithID(t, id, userID, name)
}
