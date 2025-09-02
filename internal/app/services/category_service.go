package services

import (
	"context"
	"fmt"
	"toptal/internal/app/domain"
)

type CategoryService struct {
	repo CategoryRepository
}

// NewCategoryService creates a new category service instance
func NewCategoryService(repo CategoryRepository) *CategoryService {
	return &CategoryService{
		repo: repo,
	}
}

func (s CategoryService) GetCategory(ctx context.Context, id int) (domain.Category, error) {
	if id == 0 {
		return domain.Category{}, fmt.Errorf("%w: id", domain.ErrRequired)
	}
	return s.repo.GetCategory(ctx, id)
}

func (s CategoryService) CreateCategory(ctx context.Context, category domain.Category) (domain.Category, error) {
	return s.repo.CreateCategory(ctx, category)
}

func (s CategoryService) UpdateCategory(ctx context.Context, category domain.Category) (domain.Category, error) {
	return s.repo.UpdateCategory(ctx, category)
}

func (s CategoryService) DeleteCategory(ctx context.Context, id int) error {
	if id == 0 {
		return fmt.Errorf("%w: id", domain.ErrRequired)
	}
	return s.repo.DeleteCategory(ctx, id)
}

func (s CategoryService) GetCategories(ctx context.Context) ([]domain.Category, error) {
	return s.repo.GetCategories(ctx)
}
