// internal/app/repositories/pg/category_repository.go
package pgrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"toptal/internal/app/domain"
	"toptal/internal/app/repository/models"
	"toptal/internal/pkg/pg"
)

type CategoryRepository struct {
	db *pg.DB
}

// NewCategoryRepository creates a new category repository instance
func NewCategoryRepository(db *pg.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create creates a new category
func (r *CategoryRepository) CreateCategory(ctx context.Context, category domain.Category) (domain.Category, error) {
	dbCategory := domainToCategory(category)

	var insertedCategory models.Category
	err := r.db.NewInsert().Model(&dbCategory).Returning("*").Scan(ctx, &insertedCategory)
	if err != nil {
		return domain.Category{}, fmt.Errorf("failed to insert a category: %w", err)
	}

	domainCategory, err := categoryToDomain(insertedCategory)
	if err != nil {
		return domain.Category{}, fmt.Errorf("failed to create domain category: %w", err)
	}

	return domainCategory, nil
}

// GetByID retrieves a category by ID
func (r *CategoryRepository) GetCategory(ctx context.Context, id int) (domain.Category, error) {
	var category models.Category
	err := r.db.NewSelect().Model(&category).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Category{}, domain.ErrNotFound
		}
		return domain.Category{}, fmt.Errorf("failed to get a category: %w", err)
	}

	domainCategory, err := categoryToDomain(category)
	if err != nil {
		return domain.Category{}, fmt.Errorf("failed to create domain category: %w", err)
	}

	return domainCategory, nil
}

// Update updates an existing category
func (r *CategoryRepository) UpdateCategory(ctx context.Context, category domain.Category) (domain.Category, error) {
	dbCategory := domainToCategory(category)
	dbCategory.UpdatedAt = time.Now()

	var updatedCategory models.Category
	err := r.db.NewUpdate().
		Model(&dbCategory).
		Where("id = ?", dbCategory.ID).
		ExcludeColumn("created_at").
		Returning("*").
		Scan(ctx, &updatedCategory)
	if err != nil {
		return domain.Category{}, fmt.Errorf("failed to update a category: %w", err)
	}

	domainCategory, err := categoryToDomain(updatedCategory)
	if err != nil {
		return domain.Category{}, fmt.Errorf("failed to create domain category: %w", err)
	}

	return domainCategory, nil
}

// Delete deletes a category by ID
func (r *CategoryRepository) DeleteCategory(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model((*models.Category)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete a category: %w", err)
	}

	return nil
}

// List retrieves all categories
func (r *CategoryRepository) GetCategories(ctx context.Context) ([]domain.Category, error) {
	var categories []models.Category
	err := r.db.NewSelect().Model(&categories).Order("id").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to select categories: %w", err)
	}

	var domainCategories []domain.Category
	for _, category := range categories {
		domainCategory, err := categoryToDomain(category)
		if err != nil {
			return nil, fmt.Errorf("failed to create domain category: %w", err)
		}

		domainCategories = append(domainCategories, domainCategory)
	}

	return domainCategories, nil
}
