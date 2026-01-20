package tracking

import (
	"github.com/madalinpopa/gocost-web/internal/platform/identifier"
	"github.com/madalinpopa/gocost-web/internal/platform/money"
)

type ID = identifier.ID

type Group struct {
	ID          ID
	UserID      ID
	Name        NameVO
	Description DescriptionVO
	Order       OrderVO
	Categories  []*Category
}

func NewGroup(id ID, userID ID, name NameVO, description DescriptionVO, order OrderVO) *Group {
	return &Group{
		ID:          id,
		UserID:      userID,
		Name:        name,
		Description: description,
		Order:       order,
		Categories:  make([]*Category, 0),
	}
}

func (g *Group) AddCategory(category *Category) error {
	if category.GroupID != g.ID {
		return ErrCategoryGroupMismatch
	}
	if g.hasConflictingCategory(category.Name, category.ID, category.IsRecurrent, category.StartMonth, category.EndMonth) {
		return ErrCategoryNameExists
	}
	g.Categories = append(g.Categories, category)
	return nil
}

func (g *Group) CreateCategory(id ID, name NameVO, description DescriptionVO, isRecurrent bool, startMonth Month, endMonth Month, budget money.Money) (*Category, error) {
	category, err := NewCategory(id, g.ID, name, description, isRecurrent, startMonth, endMonth, budget)
	if err != nil {
		return nil, err
	}
	if err := g.AddCategory(category); err != nil {
		return nil, err
	}
	return category, nil
}

func (g *Group) UpdateCategory(id ID, name NameVO, description DescriptionVO, isRecurrent bool, startMonth Month, endMonth Month, budget money.Money) (*Category, error) {
	var category *Category
	for _, c := range g.Categories {
		if c.ID == id {
			category = c
			break
		}
	}
	if category == nil {
		return nil, ErrCategoryNotFound
	}

	// Check for conflicts with other categories (excluding self)
	if g.hasConflictingCategory(name, id, isRecurrent, startMonth, endMonth) {
		return nil, ErrCategoryNameExists
	}

	// Validate new values
	if startMonth.IsZero() {
		return nil, ErrInvalidMonth
	}
	if !isRecurrent && !endMonth.IsZero() {
		return nil, ErrEndMonthNotAllowed
	}
	if !endMonth.IsZero() && endMonth.Before(startMonth) {
		return nil, ErrEndMonthBeforeStartMonth
	}

	category.Name = name
	category.Description = description
	category.IsRecurrent = isRecurrent
	category.StartMonth = startMonth
	category.EndMonth = endMonth
	category.Budget = budget

	return category, nil
}

func (g *Group) RemoveCategory(id ID) error {
	for i, c := range g.Categories {
		if c.ID == id {
			g.Categories = append(g.Categories[:i], g.Categories[i+1:]...)
			return nil
		}
	}
	return ErrCategoryNotFound
}

func (g *Group) CategoriesForMonth(month Month) ([]*Category, error) {
	if month.IsZero() {
		return nil, ErrInvalidMonth
	}

	categories := make([]*Category, 0, len(g.Categories))
	for _, category := range g.Categories {
		if category.IsActiveFor(month) {
			categories = append(categories, category)
		}
	}
	return categories, nil
}

func (g *Group) hasConflictingCategory(name NameVO, excludeID ID, isRecurrent bool, start Month, end Month) bool {
	// Create a temporary category object to check overlap
	// We don't care about ID/Group/Desc/Budget for overlap check
	candidate := &Category{
		Name:        name,
		IsRecurrent: isRecurrent,
		StartMonth:  start,
		EndMonth:    end,
	}

	for _, category := range g.Categories {
		if category.ID == excludeID {
			continue
		}
		if category.Name.Equals(name) {
			if category.Overlaps(candidate) {
				return true
			}
		}
	}
	return false
}

type Category struct {
	ID          ID
	GroupID     ID
	Name        NameVO
	Description DescriptionVO
	IsRecurrent bool
	StartMonth  Month
	EndMonth    Month
	Budget      money.Money
}

func NewCategory(id ID, groupID ID, name NameVO, description DescriptionVO, isRecurrent bool, startMonth Month, endMonth Month, budget money.Money) (*Category, error) {
	if startMonth.IsZero() {
		return nil, ErrInvalidMonth
	}
	if !isRecurrent && !endMonth.IsZero() {
		return nil, ErrEndMonthNotAllowed
	}
	if !endMonth.IsZero() && endMonth.Before(startMonth) {
		return nil, ErrEndMonthBeforeStartMonth
	}

	return &Category{
		ID:          id,
		GroupID:     groupID,
		Name:        name,
		Description: description,
		IsRecurrent: isRecurrent,
		StartMonth:  startMonth,
		EndMonth:    endMonth,
		Budget:      budget,
	}, nil
}

func (c *Category) IsActiveFor(month Month) bool {
	if month.IsZero() || c.StartMonth.IsZero() {
		return false
	}
	if month.Before(c.StartMonth) {
		return false
	}
	if !c.IsRecurrent {
		return month.Equals(c.StartMonth)
	}
	if c.EndMonth.IsZero() {
		return true
	}
	return !c.EndMonth.Before(month)
}

func (c *Category) Overlaps(other *Category) bool {
	// Define intervals [StartA, EndA] and [StartB, EndB]
	startA := c.StartMonth
	endA := c.EndMonth
	if !c.IsRecurrent {
		endA = c.StartMonth
	}

	startB := other.StartMonth
	endB := other.EndMonth
	if !other.IsRecurrent {
		endB = other.StartMonth
	}

	// Check 1: StartA <= EndB
	// If EndB is infinity (zero), then StartA <= Infinity is always true.
	// We only check if EndB is NOT zero.
	if !endB.IsZero() {
		if endB.Before(startA) {
			return false // EndB < StartA, so no overlap
		}
	}

	// Check 2: StartB <= EndA
	// If EndA is infinity (zero), then StartB <= Infinity is always true.
	if !endA.IsZero() {
		if endA.Before(startB) {
			return false // EndA < StartB, so no overlap
		}
	}

	return true
}
