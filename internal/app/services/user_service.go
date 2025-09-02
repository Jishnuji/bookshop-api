package services

import (
	"context"
	"toptal/internal/app/domain"
)

// UserService is a user service instance
type UserService struct {
	repo UserRepository
}

// NewUserService creates a new user service instance
func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (u UserService) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	return u.repo.GetUserByEmail(ctx, email)
}

func (u UserService) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	return u.repo.CreateUser(ctx, user)
}

func (u UserService) GetUser(ctx context.Context, id int) (domain.User, error) {
	return u.repo.GetUser(ctx, id)
}
