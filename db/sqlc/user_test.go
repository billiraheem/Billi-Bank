package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/billiraheem/Billi-Bank/utils"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	hasedPassword, err := utils.HashPassword(utils.RandomString(6))
	require.NoError(t, err)

	args := CreateUserParams{
		Username:    utils.RandomOwner(),
		HashedPassword:  hasedPassword,
		Fullname: utils.RandomOwner(),
		Email: utils.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, args.Username, user.Username)
	require.Equal(t, args.HashedPassword, user.HashedPassword)
	require.Equal(t, args.Fullname, user.Fullname)
	require.Equal(t, args.Email, user.Email)

	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testQueries.GetUser(context.Background(), user1.Username)

	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.Fullname, user2.Fullname)
	require.Equal(t, user1.Email, user2.Email)
	require.WithinDuration(t, user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestUpdateUserOnlyFullname(t *testing.T) {
	oldUser := createRandomUser(t)

	newFullname := utils.RandomOwner()
	arg := UpdateUserParams {
		Username: oldUser.Username,
		Fullname: sql.NullString{
			String: newFullname,
			Valid: true,
		},
	}
	updatedUser, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, newFullname, updatedUser.Fullname)
	require.NotEqual(t, oldUser.Fullname, updatedUser.Fullname)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.WithinDuration(t, oldUser.PasswordChangedAt, updatedUser.PasswordChangedAt, time.Second)
	require.WithinDuration(t, oldUser.CreatedAt, updatedUser.CreatedAt, time.Second)
}

func TestUpdateUserOnlyEmail(t *testing.T) {
	oldUser := createRandomUser(t)
	
	newEmail := utils.RandomEmail()
	arg := UpdateUserParams {
		Username: oldUser.Username,
		Email: sql.NullString{
			String: newEmail,
			Valid: true,
		},
	}
	updatedUser, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, newEmail, updatedUser.Email)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.Fullname, updatedUser.Fullname)
	require.WithinDuration(t, oldUser.PasswordChangedAt, updatedUser.PasswordChangedAt, time.Second)
	require.WithinDuration(t, oldUser.CreatedAt, updatedUser.CreatedAt, time.Second)
}

func TestUpdateUserOnlyPassword(t *testing.T) {
	oldUser := createRandomUser(t)

	newHashedPassword, err := utils.HashPassword(utils.RandomString(6))
	require.NoError(t, err)

	arg := UpdateUserParams {
		Username: oldUser.Username,
		HashedPassword: sql.NullString{
			String: newHashedPassword,
			Valid: true,
		},
	}
	updatedUser, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, oldUser.Fullname, updatedUser.Fullname)
	require.WithinDuration(t, oldUser.PasswordChangedAt, updatedUser.PasswordChangedAt, time.Second)
	require.WithinDuration(t, oldUser.CreatedAt, updatedUser.CreatedAt, time.Second)
}

func TestUpdateUser(t *testing.T) {
	oldUser := createRandomUser(t)

	newFullname := utils.RandomOwner()
	newEmail := utils.RandomEmail()
	newHashedPassword, err := utils.HashPassword(utils.RandomString(6))
	require.NoError(t, err)

	arg := UpdateUserParams {
		Username: oldUser.Username,
		Fullname: sql.NullString{
			String: newFullname,
			Valid: true,
		},
		Email: sql.NullString{
			String: newEmail,
			Valid: true,
		},
		HashedPassword: sql.NullString{
			String: newHashedPassword,
			Valid: true,
		},
	}
	updatedUser, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)

	require.Equal(t, newFullname, updatedUser.Fullname)
	require.NotEqual(t, oldUser.Fullname, updatedUser.Fullname)
	require.Equal(t, newEmail, updatedUser.Email)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.WithinDuration(t, oldUser.PasswordChangedAt, updatedUser.PasswordChangedAt, time.Second)
	require.WithinDuration(t, oldUser.CreatedAt, updatedUser.CreatedAt, time.Second)
}