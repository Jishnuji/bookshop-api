package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser_Success(t *testing.T) {
	// Arrange
	userData := NewUserData{
		ID:       1,
		Email:    "john.doe@example.com",
		Password: "password123",
		Admin:    false,
	}

	// Act
	user, err := NewUser(userData)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, user.ID())
	assert.Equal(t, "john.doe@example.com", user.Email())
	assert.Equal(t, "password123", user.Password())
	assert.False(t, user.Admin())
}

func TestNewUser_EmptyEmail(t *testing.T) {
	// Arrange
	userData := NewUserData{
		ID:       1,
		Email:    "", // empty email
		Password: "password123",
		Admin:    false,
	}

	// Act
	user, err := NewUser(userData)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRequired)
	assert.Contains(t, err.Error(), "email")
	assert.Equal(t, User{}, user)
}

func TestNewUser_EmptyPassword(t *testing.T) {
	// Arrange
	userData := NewUserData{
		ID:       1,
		Email:    "john.doe@example.com",
		Password: "", // empty password
		Admin:    false,
	}

	// Act
	user, err := NewUser(userData)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRequired)
	assert.Contains(t, err.Error(), "password")
	assert.Equal(t, User{}, user)
}

func TestNewUser_EmptyEmailAndPassword(t *testing.T) {
	// Arrange
	userData := NewUserData{
		ID:       1,
		Email:    "", // empty email
		Password: "", // empty password
		Admin:    false,
	}

	// Act
	user, err := NewUser(userData)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRequired)
	// First error - email (validation order)
	assert.Contains(t, err.Error(), "email")
	assert.Equal(t, User{}, user)
}

func TestNewUser_AdminUser(t *testing.T) {
	// Arrange
	userData := NewUserData{
		ID:       1,
		Email:    "admin@example.com",
		Password: "adminpass123",
		Admin:    true, // admin flag
	}

	// Act
	user, err := NewUser(userData)

	// Assert
	require.NoError(t, err)
	assert.True(t, user.Admin())
}

func TestNewUserFromToken_Success(t *testing.T) {
	// Arrange
	userData := NewUserData{
		ID:    1,
		Email: "john.doe@example.com",
		Admin: false,
		// Password is not required for token
	}

	// Act
	user, err := NewUserFromToken(userData)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, user.ID())
	assert.Equal(t, "john.doe@example.com", user.Email())
	assert.False(t, user.Admin())
	assert.Empty(t, user.Password()) // password is empty for token
}

func TestNewUserFromToken_ZeroID(t *testing.T) {
	// Arrange
	userData := NewUserData{
		ID:    0, // zero ID
		Email: "john.doe@example.com",
		Admin: false,
	}

	// Act
	user, err := NewUserFromToken(userData)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRequired)
	assert.Contains(t, err.Error(), "id")
	assert.Equal(t, User{}, user)
}

func TestNewUserFromToken_EmptyEmail(t *testing.T) {
	// Arrange
	userData := NewUserData{
		ID:    1,
		Email: "", // empty email
		Admin: false,
	}

	// Act
	user, err := NewUserFromToken(userData)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRequired)
	assert.Contains(t, err.Error(), "email")
	assert.Equal(t, User{}, user)
}
