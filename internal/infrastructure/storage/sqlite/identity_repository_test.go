package sqlite_test

import (
	"context"
	"testing"

	"github.com/madalinpopa/gocost-web/internal/domain/identity"
	"github.com/madalinpopa/gocost-web/internal/infrastructure/storage/sqlite"
	"github.com/madalinpopa/gocost-web/internal/shared/identifier"
	"github.com/stretchr/testify/assert"
)

func TestSQLiteUserRepository(t *testing.T) {
	repo := sqlite.NewSQLiteUserRepository(testDB)
	ctx := context.Background()

	t.Run("Save_Success", func(t *testing.T) {
		user := createRandomUser(t)
		err := repo.Save(ctx, *user)
		assert.NoError(t, err)

		exists, err := repo.ExistsByEmail(ctx, user.Email)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("FindByID_Success", func(t *testing.T) {
		user := createRandomUser(t)
		err := repo.Save(ctx, *user)
		assert.NoError(t, err)

		foundUser, err := repo.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, user.ID.String(), foundUser.ID.String())
		assert.Equal(t, user.Email.Value(), foundUser.Email.Value())
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		randomID, _ := identifier.NewID()
		_, err := repo.FindByID(ctx, randomID)
		assert.ErrorIs(t, err, identity.ErrUserNotFound)
	})

	t.Run("FindByEmail_Success", func(t *testing.T) {
		user := createRandomUser(t)
		err := repo.Save(ctx, *user)
		assert.NoError(t, err)

		foundUser, err := repo.FindByEmail(ctx, user.Email)
		assert.NoError(t, err)
		assert.Equal(t, user.ID.String(), foundUser.ID.String())
	})

	t.Run("FindByEmail_NotFound", func(t *testing.T) {
		randomEmail, _ := identity.NewEmailVO("random@example.com")
		_, err := repo.FindByEmail(ctx, randomEmail)
		assert.ErrorIs(t, err, identity.ErrUserNotFound)
	})

	t.Run("FindByUsername_Success", func(t *testing.T) {
		user := createRandomUser(t)
		err := repo.Save(ctx, *user)
		assert.NoError(t, err)

		foundUser, err := repo.FindByUsername(ctx, user.Username)
		assert.NoError(t, err)
		assert.Equal(t, user.ID.String(), foundUser.ID.String())
	})

	t.Run("FindByUsername_NotFound", func(t *testing.T) {
		randomUsername, _ := identity.NewUsernameVO("randomuser")
		_, err := repo.FindByUsername(ctx, randomUsername)
		assert.ErrorIs(t, err, identity.ErrUserNotFound)
	})

	t.Run("Save_DuplicateEmail", func(t *testing.T) {
		user1 := createRandomUser(t)
		err := repo.Save(ctx, *user1)
		assert.NoError(t, err)

		user2 := createRandomUser(t)
		duplicateUser := identity.NewUser(user2.ID, user2.Username, user1.Email, user2.Password)

		err = repo.Save(ctx, *duplicateUser)
		assert.ErrorIs(t, err, identity.ErrUserAlreadyExists)
	})

	t.Run("Save_DuplicateUsername", func(t *testing.T) {
		user1 := createRandomUser(t)
		err := repo.Save(ctx, *user1)
		assert.NoError(t, err)

		user2 := createRandomUser(t)
		duplicateUser := identity.NewUser(user2.ID, user1.Username, user2.Email, user2.Password)

		err = repo.Save(ctx, *duplicateUser)
		assert.ErrorIs(t, err, identity.ErrUserAlreadyExists)
	})

	t.Run("ExistsByUsername", func(t *testing.T) {
		user := createRandomUser(t)
		err := repo.Save(ctx, *user)
		assert.NoError(t, err)

		exists, err := repo.ExistsByUsername(ctx, user.Username)
		assert.NoError(t, err)
		assert.True(t, exists)

		otherUsername, _ := identity.NewUsernameVO("nonexistent_user")
		exists, err = repo.ExistsByUsername(ctx, otherUsername)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Update_User", func(t *testing.T) {
		user := createRandomUser(t)
		err := repo.Save(ctx, *user)
		assert.NoError(t, err)

		newUsername, _ := identity.NewUsernameVO("updated" + user.Username.Value())
		updatedUser := identity.NewUser(user.ID, newUsername, user.Email, user.Password)

		err = repo.Save(ctx, *updatedUser)
		assert.NoError(t, err)

		foundUser, err := repo.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, newUsername.Value(), foundUser.Username.Value())
	})
}
