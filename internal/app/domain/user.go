package domain

import (
	"fmt"
)

type User struct {
	id       int
	email    string
	password string
	admin    bool
}

type NewUserData struct {
	ID       int
	Email    string
	Password string
	Admin    bool
}

// NewUser constructs a User from the provided data.
func NewUser(data NewUserData) (User, error) {
	if err := validateUserData(data); err != nil {
		return User{}, fmt.Errorf("faild user data validation: %w", err)
	}

	return User{
		id:       data.ID,
		email:    data.Email,
		password: data.Password,
		admin:    data.Admin,
	}, nil
}

func NewUserFromToken(data NewUserData) (User, error) {
	if data.ID == 0 {
		return User{}, fmt.Errorf("%w: id", ErrRequired)
	}
	if data.Email == "" {
		return User{}, fmt.Errorf("%w: email", ErrRequired)
	}

	return User{
		id:    data.ID,
		email: data.Email,
		admin: data.Admin,
	}, nil

}

// validateUserData checks that required fields (email, password) are provided.
func validateUserData(data NewUserData) error {
	if data.Email == "" {
		return fmt.Errorf("%w: email", ErrRequired)
	}
	if data.Password == "" {
		return fmt.Errorf("%w: password", ErrRequired)
	}
	return nil
}

// ID returns the user identifier.
func (u User) ID() int {
	return u.id
}

// Email returns the user's email.
func (u User) Email() string {
	return u.email
}

// Password returns the stored password value (e.g., hashed).
func (u User) Password() string {
	return u.password
}

// Admin reports whether the user has admin privileges.
func (u User) Admin() bool {
	return u.admin
}
