package tracking

import "context"

// GroupRepository defines the contract for group aggregate persistence.
type GroupRepository interface {
	Save(ctx context.Context, group Group) error
	FindByID(ctx context.Context, id ID) (Group, error)
	FindByUserID(ctx context.Context, userID ID) ([]Group, error)
	FindGroupByCategoryID(ctx context.Context, categoryID ID) (Group, error)
	Delete(ctx context.Context, id ID) error
	DeleteCategory(ctx context.Context, id ID) error
}
