package tracking

import (
	"github.com/madalinpopa/gocost-web/internal/shared/identifier"
	"github.com/madalinpopa/gocost-web/internal/shared/money"
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
	if g.hasCategoryName(category.Name) {
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

	if !category.Name.Equals(name) {
		if g.hasCategoryName(name) {
			return nil, ErrCategoryNameExists
		}
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

func (g *Group) hasCategoryName(name NameVO) bool {
	for _, category := range g.Categories {
		if category.Name.Equals(name) {
			return true
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
