package identity

import "context"

// UserRepository defines the contract for user data persistence operations.
type UserRepository interface {
	Save(ctx context.Context, user User) error
	FindByID(ctx context.Context, id ID) (User, error)
	FindByEmail(ctx context.Context, email EmailVO) (User, error)
	FindByUsername(ctx context.Context, username UsernameVO) (User, error)
	ExistsByEmail(ctx context.Context, email EmailVO) (bool, error)
	ExistsByUsername(ctx context.Context, username UsernameVO) (bool, error)
}
