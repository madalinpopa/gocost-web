package tracking

import (
	"testing"
	"time"

	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGroup(t *testing.T) {
	t.Run("creates valid group", func(t *testing.T) {
		// Arrange
		id, _ := identifier.NewID()
		userID, _ := identifier.NewID()
		name, _ := NewNameVO("Personal")
		description, _ := NewDescriptionVO("Personal expenses")
		order, _ := NewOrderVO(0)

		// Act
		group := NewGroup(id, userID, name, description, order)

		// Assert
		assert.NotNil(t, group)
		assert.Equal(t, id, group.ID)
		assert.Equal(t, userID, group.UserID)
		assert.Equal(t, name, group.Name)
		assert.Equal(t, description, group.Description)
		assert.Equal(t, order, group.Order)
		assert.Empty(t, group.Categories)
	})

	t.Run("add category to group", func(t *testing.T) {
		// Arrange
		groupID, _ := identifier.NewID()
		userID, _ := identifier.NewID()
		groupName, _ := NewNameVO("Personal")
		groupDesc, _ := NewDescriptionVO("Personal expenses")
		group := NewGroup(groupID, userID, groupName, groupDesc, mustOrder(t, 0))

		catID, _ := identifier.NewID()
		catName, _ := NewNameVO("Food")
		catDesc, _ := NewDescriptionVO("Food expenses")
		startMonth, err := NewMonth(2024, time.January)
		require.NoError(t, err)
		category, err := NewCategory(catID, groupID, catName, catDesc, false, startMonth, Month{}, money.Money{})
		require.NoError(t, err)

		// Act
		err = group.AddCategory(category)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, group.Categories, 1)
		assert.Equal(t, category, group.Categories[0])
	})

	t.Run("rejects duplicate category name", func(t *testing.T) {
		groupID, _ := identifier.NewID()
		userID, _ := identifier.NewID()
		groupName, _ := NewNameVO("Personal")
		groupDesc, _ := NewDescriptionVO("Personal expenses")
		group := NewGroup(groupID, userID, groupName, groupDesc, mustOrder(t, 0))

		catName, _ := NewNameVO("Food")
		catDesc, _ := NewDescriptionVO("Food expenses")
		startMonth, err := NewMonth(2024, time.January)
		require.NoError(t, err)

		firstCatID, err := identifier.NewID()
		require.NoError(t, err)
		firstCat, err := NewCategory(firstCatID, groupID, catName, catDesc, false, startMonth, Month{}, money.Money{})
		require.NoError(t, err)
		require.NoError(t, group.AddCategory(firstCat))

		secondCatID, err := identifier.NewID()
		require.NoError(t, err)
		secondCat, err := NewCategory(secondCatID, groupID, catName, catDesc, false, startMonth, Month{}, money.Money{})
		require.NoError(t, err)

		err = group.AddCategory(secondCat)
		assert.ErrorIs(t, err, ErrCategoryNameExists)
	})

	t.Run("rejects category with mismatched group", func(t *testing.T) {
		groupID, _ := identifier.NewID()
		userID, _ := identifier.NewID()
		groupName, _ := NewNameVO("Personal")
		groupDesc, _ := NewDescriptionVO("Personal expenses")
		group := NewGroup(groupID, userID, groupName, groupDesc, mustOrder(t, 0))

		otherGroupID, _ := identifier.NewID()
		catName, _ := NewNameVO("Food")
		catDesc, _ := NewDescriptionVO("Food expenses")
		startMonth, err := NewMonth(2024, time.January)
		require.NoError(t, err)

		category, err := NewCategory(otherGroupID, otherGroupID, catName, catDesc, false, startMonth, Month{}, money.Money{})
		require.NoError(t, err)

		err = group.AddCategory(category)
		assert.ErrorIs(t, err, ErrCategoryGroupMismatch)
	})
}

func TestNewCategory(t *testing.T) {
	t.Run("creates valid category", func(t *testing.T) {
		// Arrange
		id, _ := identifier.NewID()
		groupID, _ := identifier.NewID()
		name, _ := NewNameVO("Food")
		description, _ := NewDescriptionVO("Food and drinks")
		startMonth, err := NewMonth(2024, time.January)
		require.NoError(t, err)

		// Act
		category, err := NewCategory(id, groupID, name, description, false, startMonth, Month{}, money.Money{})

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, category)
		assert.Equal(t, id, category.ID)
		assert.Equal(t, groupID, category.GroupID)
		assert.Equal(t, name, category.Name)
		assert.Equal(t, description, category.Description)
		assert.False(t, category.IsRecurrent)
		assert.Equal(t, startMonth, category.StartMonth)
		assert.True(t, category.EndMonth.IsZero())
	})

	t.Run("rejects end month for non-recurrent category", func(t *testing.T) {
		id, _ := identifier.NewID()
		groupID, _ := identifier.NewID()
		name, _ := NewNameVO("Food")
		description, _ := NewDescriptionVO("Food and drinks")
		startMonth, err := NewMonth(2024, time.January)
		require.NoError(t, err)
		endMonth, err := NewMonth(2024, time.February)
		require.NoError(t, err)

		category, err := NewCategory(id, groupID, name, description, false, startMonth, endMonth, money.Money{})
		assert.ErrorIs(t, err, ErrEndMonthNotAllowed)
		assert.Nil(t, category)
	})

	t.Run("rejects end month before start month", func(t *testing.T) {
		id, _ := identifier.NewID()
		groupID, _ := identifier.NewID()
		name, _ := NewNameVO("Food")
		description, _ := NewDescriptionVO("Food and drinks")
		startMonth, err := NewMonth(2024, time.March)
		require.NoError(t, err)
		endMonth, err := NewMonth(2024, time.January)
		require.NoError(t, err)

		category, err := NewCategory(id, groupID, name, description, true, startMonth, endMonth, money.Money{})
		assert.ErrorIs(t, err, ErrEndMonthBeforeStartMonth)
		assert.Nil(t, category)
	})
}

func TestGroupCategoriesForMonth(t *testing.T) {
	groupID, _ := identifier.NewID()
	userID, _ := identifier.NewID()
	groupName, _ := NewNameVO("Personal")
	groupDesc, _ := NewDescriptionVO("Personal expenses")
	group := NewGroup(groupID, userID, groupName, groupDesc, mustOrder(t, 0))

	startJan, err := NewMonth(2024, time.January)
	require.NoError(t, err)
	endMar, err := NewMonth(2024, time.March)
	require.NoError(t, err)

	nonRecurrentID, err := identifier.NewID()
	require.NoError(t, err)
	nonRecurrent, err := NewCategory(nonRecurrentID, groupID, mustName(t, "One-off"), mustDesc(t, "One-off"), false, startJan, Month{}, money.Money{})
	require.NoError(t, err)
	require.NoError(t, group.AddCategory(nonRecurrent))

	recurrentNoEndID, err := identifier.NewID()
	require.NoError(t, err)
	recurrentNoEnd, err := NewCategory(recurrentNoEndID, groupID, mustName(t, "Rent"), mustDesc(t, "Monthly rent"), true, startJan, Month{}, money.Money{})
	require.NoError(t, err)
	require.NoError(t, group.AddCategory(recurrentNoEnd))

	recurrentWithEndID, err := identifier.NewID()
	require.NoError(t, err)
	recurrentWithEnd, err := NewCategory(recurrentWithEndID, groupID, mustName(t, "Promo"), mustDesc(t, "Limited"), true, startJan, endMar, money.Money{})
	require.NoError(t, err)
	require.NoError(t, group.AddCategory(recurrentWithEnd))

	feb, err := NewMonth(2024, time.February)
	require.NoError(t, err)
	categories, err := group.CategoriesForMonth(feb)
	require.NoError(t, err)
	assert.ElementsMatch(t, []*Category{recurrentNoEnd, recurrentWithEnd}, categories)

	apr, err := NewMonth(2024, time.April)
	require.NoError(t, err)
	categories, err = group.CategoriesForMonth(apr)
	require.NoError(t, err)
	assert.Len(t, categories, 1)
	assert.Equal(t, recurrentNoEnd, categories[0])
}

func TestGroup_UpdateCategory(t *testing.T) {
	groupID, _ := identifier.NewID()
	userID, _ := identifier.NewID()
	groupName, _ := NewNameVO("Group")
	groupDesc, _ := NewDescriptionVO("Desc")
	group := NewGroup(groupID, userID, groupName, groupDesc, mustOrder(t, 0))

	catID, _ := identifier.NewID()
	startMonth, _ := NewMonth(2024, time.January)

	// Add initial category
	_, err := group.CreateCategory(catID, mustName(t, "Old Name"), mustDesc(t, "Old Desc"), false, startMonth, Month{}, money.Money{})
	require.NoError(t, err)

	t.Run("updates category successfully", func(t *testing.T) {
		newName := mustName(t, "New Name")
		newDesc := mustDesc(t, "New Desc")
		newStart, _ := NewMonth(2024, time.February)

		updated, err := group.UpdateCategory(catID, newName, newDesc, false, newStart, Month{}, money.Money{})
		require.NoError(t, err)
		assert.Equal(t, newName, updated.Name)
		assert.Equal(t, newDesc, updated.Description)
		assert.Equal(t, newStart, updated.StartMonth)

		// Verify changes persisted in group
		found := false
		for _, c := range group.Categories {
			if c.ID == catID {
				assert.Equal(t, newName, c.Name)
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("returns error when category not found", func(t *testing.T) {
		otherID, _ := identifier.NewID()
		_, err := group.UpdateCategory(otherID, mustName(t, "Name"), mustDesc(t, "Desc"), false, startMonth, Month{}, money.Money{})
		assert.ErrorIs(t, err, ErrCategoryNotFound)
	})

	t.Run("returns error on duplicate name", func(t *testing.T) {
		otherCatID, _ := identifier.NewID()
		otherName := mustName(t, "Other")
		_, err := group.CreateCategory(otherCatID, otherName, mustDesc(t, "Desc"), false, startMonth, Month{}, money.Money{})
		require.NoError(t, err)

		_, err = group.UpdateCategory(catID, otherName, mustDesc(t, "New Desc"), false, startMonth, Month{}, money.Money{})
		assert.ErrorIs(t, err, ErrCategoryNameExists)
	})
}

func TestGroup_CreateCategory(t *testing.T) {
	t.Run("creates category successfully", func(t *testing.T) {
		// Arrange
		groupID, _ := identifier.NewID()
		userID, _ := identifier.NewID()
		groupName, _ := NewNameVO("Group")
		groupDesc, _ := NewDescriptionVO("Desc")
		group := NewGroup(groupID, userID, groupName, groupDesc, mustOrder(t, 0))

		catID, _ := identifier.NewID()
		catName := mustName(t, "Food")
		catDesc := mustDesc(t, "Food expenses")
		startMonth, _ := NewMonth(2024, time.January)
		budget, _ := money.New(500)

		// Act
		category, err := group.CreateCategory(catID, catName, catDesc, false, startMonth, Month{}, budget)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, category)
		assert.Equal(t, catID, category.ID)
		assert.Equal(t, groupID, category.GroupID)
		assert.Equal(t, catName, category.Name)
		assert.Equal(t, catDesc, category.Description)
		assert.False(t, category.IsRecurrent)
		assert.Equal(t, startMonth, category.StartMonth)
		assert.True(t, category.EndMonth.IsZero())
		assert.Equal(t, budget, category.Budget)
		assert.Len(t, group.Categories, 1)
		assert.Equal(t, category, group.Categories[0])
	})

	t.Run("rejects duplicate category name", func(t *testing.T) {
		// Arrange
		groupID, _ := identifier.NewID()
		userID, _ := identifier.NewID()
		groupName, _ := NewNameVO("Group")
		groupDesc, _ := NewDescriptionVO("Desc")
		group := NewGroup(groupID, userID, groupName, groupDesc, mustOrder(t, 0))

		catName := mustName(t, "Food")
		catDesc := mustDesc(t, "Desc")
		startMonth, _ := NewMonth(2024, time.January)

		firstCatID, _ := identifier.NewID()
		_, err := group.CreateCategory(firstCatID, catName, catDesc, false, startMonth, Month{}, money.Money{})
		require.NoError(t, err)

		// Act
		secondCatID, _ := identifier.NewID()
		_, err = group.CreateCategory(secondCatID, catName, catDesc, false, startMonth, Month{}, money.Money{})

		// Assert
		assert.ErrorIs(t, err, ErrCategoryNameExists)
		assert.Len(t, group.Categories, 1)
	})

	t.Run("rejects invalid category data", func(t *testing.T) {
		// Arrange
		groupID, _ := identifier.NewID()
		userID, _ := identifier.NewID()
		groupName, _ := NewNameVO("Group")
		groupDesc, _ := NewDescriptionVO("Desc")
		group := NewGroup(groupID, userID, groupName, groupDesc, mustOrder(t, 0))

		catID, _ := identifier.NewID()
		catName := mustName(t, "Food")
		catDesc := mustDesc(t, "Desc")

		// Act - missing start month
		_, err := group.CreateCategory(catID, catName, catDesc, false, Month{}, Month{}, money.Money{})

		// Assert
		assert.ErrorIs(t, err, ErrInvalidMonth)
		assert.Empty(t, group.Categories)
	})
}

func TestGroup_RemoveCategory(t *testing.T) {
	groupID, _ := identifier.NewID()
	userID, _ := identifier.NewID()
	groupName, _ := NewNameVO("Group")
	groupDesc, _ := NewDescriptionVO("Desc")
	group := NewGroup(groupID, userID, groupName, groupDesc, mustOrder(t, 0))

	catID, _ := identifier.NewID()
	startMonth, _ := NewMonth(2024, time.January)

	_, err := group.CreateCategory(catID, mustName(t, "To Delete"), mustDesc(t, "Desc"), false, startMonth, Month{}, money.Money{})
	require.NoError(t, err)

	t.Run("removes category successfully", func(t *testing.T) {
		err := group.RemoveCategory(catID)
		require.NoError(t, err)
		assert.Empty(t, group.Categories)
	})

	t.Run("returns error when category not found", func(t *testing.T) {
		err := group.RemoveCategory(catID) // Already removed
		assert.ErrorIs(t, err, ErrCategoryNotFound)
	})
}

func mustName(t *testing.T, value string) NameVO {
	t.Helper()
	name, err := NewNameVO(value)
	require.NoError(t, err)
	return name
}

func mustDesc(t *testing.T, value string) DescriptionVO {
	t.Helper()
	desc, err := NewDescriptionVO(value)
	require.NoError(t, err)
	return desc
}

func mustOrder(t *testing.T, value int) OrderVO {
	t.Helper()
	order, err := NewOrderVO(value)
	require.NoError(t, err)
	return order
}

func TestCategory_IsActiveFor(t *testing.T) {
	groupID, _ := identifier.NewID()
	catID, _ := identifier.NewID()

	t.Run("non-recurrent category is active only for start month", func(t *testing.T) {
		// Arrange
		startMonth, _ := NewMonth(2024, time.January)
		category, err := NewCategory(catID, groupID, mustName(t, "One-off"), mustDesc(t, ""), false, startMonth, Month{}, money.Money{})
		require.NoError(t, err)

		jan, _ := NewMonth(2024, time.January)
		feb, _ := NewMonth(2024, time.February)

		// Act & Assert
		assert.True(t, category.IsActiveFor(jan))
		assert.False(t, category.IsActiveFor(feb))
	})

	t.Run("recurrent category without end month is active from start month onwards", func(t *testing.T) {
		// Arrange
		startMonth, _ := NewMonth(2024, time.January)
		category, err := NewCategory(catID, groupID, mustName(t, "Rent"), mustDesc(t, ""), true, startMonth, Month{}, money.Money{})
		require.NoError(t, err)

		dec2023, _ := NewMonth(2023, time.December)
		jan2024, _ := NewMonth(2024, time.January)
		feb2024, _ := NewMonth(2024, time.February)
		dec2024, _ := NewMonth(2024, time.December)

		// Act & Assert
		assert.False(t, category.IsActiveFor(dec2023))
		assert.True(t, category.IsActiveFor(jan2024))
		assert.True(t, category.IsActiveFor(feb2024))
		assert.True(t, category.IsActiveFor(dec2024))
	})

	t.Run("recurrent category with end month is active between start and end", func(t *testing.T) {
		// Arrange
		startMonth, _ := NewMonth(2024, time.January)
		endMonth, _ := NewMonth(2024, time.March)
		category, err := NewCategory(catID, groupID, mustName(t, "Promo"), mustDesc(t, ""), true, startMonth, endMonth, money.Money{})
		require.NoError(t, err)

		dec2023, _ := NewMonth(2023, time.December)
		jan2024, _ := NewMonth(2024, time.January)
		feb2024, _ := NewMonth(2024, time.February)
		mar2024, _ := NewMonth(2024, time.March)
		apr2024, _ := NewMonth(2024, time.April)

		// Act & Assert
		assert.False(t, category.IsActiveFor(dec2023))
		assert.True(t, category.IsActiveFor(jan2024))
		assert.True(t, category.IsActiveFor(feb2024))
		assert.True(t, category.IsActiveFor(mar2024))
		assert.False(t, category.IsActiveFor(apr2024))
	})

	t.Run("returns false for zero month", func(t *testing.T) {
		// Arrange
		startMonth, _ := NewMonth(2024, time.January)
		category, err := NewCategory(catID, groupID, mustName(t, "Test"), mustDesc(t, ""), false, startMonth, Month{}, money.Money{})
		require.NoError(t, err)

		// Act & Assert
		assert.False(t, category.IsActiveFor(Month{}))
	})

	t.Run("returns false when category has zero start month", func(t *testing.T) {
		// Arrange - manually create category with zero start month (bypassing constructor)
		category := &Category{
			ID:          catID,
			GroupID:     groupID,
			Name:        mustName(t, "Test"),
			Description: mustDesc(t, ""),
			StartMonth:  Month{},
		}

		queryMonth, _ := NewMonth(2024, time.January)

		// Act & Assert
		assert.False(t, category.IsActiveFor(queryMonth))
	})
}

func TestGroup_AddCategory_Overlap(t *testing.T) {
	groupID, _ := identifier.NewID()
	userID, _ := identifier.NewID()
	name, _ := NewNameVO("Group")
	desc, _ := NewDescriptionVO("Desc")
	order, _ := NewOrderVO(1)
	group := NewGroup(groupID, userID, name, desc, order)

	catName, _ := NewNameVO("Food")
	catDesc, _ := NewDescriptionVO("Desc")
	budget, _ := money.New(100)
	jan := NewMonthFromTime(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
	feb := NewMonthFromTime(time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC))
	mar := NewMonthFromTime(time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC))

	// 1. Add Jan-Feb
	id1, _ := identifier.NewID()
	cat1, _ := NewCategory(id1, groupID, catName, catDesc, true, jan, feb, budget)
	err := group.AddCategory(cat1)
	require.NoError(t, err)

	// 2. Add Mar-Inf (Should Succeed)
	id2, _ := identifier.NewID()
	cat2, _ := NewCategory(id2, groupID, catName, catDesc, true, mar, Month{}, budget)
	err = group.AddCategory(cat2)
	assert.NoError(t, err, "Should allow non-overlapping category with same name")

	// 3. Add Jan-Inf (Should Fail - Overlaps with both)
	id3, _ := identifier.NewID()
	cat3, _ := NewCategory(id3, groupID, catName, catDesc, true, jan, Month{}, budget)
	err = group.AddCategory(cat3)
	assert.ErrorIs(t, err, ErrCategoryNameExists)

	// 4. Add Feb-Mar (Should Fail - Overlaps with both)
	id4, _ := identifier.NewID()
	cat4, _ := NewCategory(id4, groupID, catName, catDesc, true, feb, mar, budget)
	err = group.AddCategory(cat4)
	assert.ErrorIs(t, err, ErrCategoryNameExists)
}
