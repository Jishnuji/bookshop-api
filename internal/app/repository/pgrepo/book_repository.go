// internal/app/repositories/pg/book_repository.go
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

	"github.com/uptrace/bun"
)

type BookRepository struct {
	db *pg.DB
}

// NewBookRepository creates a new book repository instance
func NewBookRepository(db *pg.DB) *BookRepository {
	return &BookRepository{db: db}
}

// Create creates a new book
func (r *BookRepository) CreateBook(ctx context.Context, book domain.Book) (domain.Book, error) {
	dbBook := domainToBook(book)

	var insertedBook models.Book
	err := r.db.NewInsert().Model(&dbBook).Returning("*").Scan(ctx, &insertedBook)
	if err != nil {
		return domain.Book{}, fmt.Errorf("failed to insert a book: %w", err)
	}

	domainBook, err := bookToDomain(insertedBook)
	if err != nil {
		return domain.Book{}, fmt.Errorf("failed to create domain book: %w", err)
	}

	return domainBook, nil
}

// GetByID retrieves a book by ID
func (r *BookRepository) GetBook(ctx context.Context, id int) (domain.Book, error) {
	var book models.Book
	err := r.db.NewSelect().Model(&book).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Book{}, domain.ErrNotFound
		}
		return domain.Book{}, fmt.Errorf("failed to get a book: %w", err)
	}

	domainBook, err := bookToDomain(book)
	if err != nil {
		return domain.Book{}, fmt.Errorf("failed to create domain book: %w", err)
	}

	return domainBook, nil
}

// Update updates an existing book
func (r *BookRepository) UpdateBook(ctx context.Context, book domain.Book) (domain.Book, error) {
	dbBook := domainToBook(book)
	dbBook.UpdatedAt = time.Now()

	var updatedBook models.Book
	err := r.db.NewUpdate().
		Model(&dbBook).
		Where("id = ?", dbBook.ID).
		ExcludeColumn("created_at", "stock").
		Returning("*").
		Scan(ctx, &updatedBook)
	if err != nil {
		return domain.Book{}, fmt.Errorf("failed to update a book: %w", err)
	}

	domainBook, err := bookToDomain(updatedBook)
	if err != nil {
		return domain.Book{}, fmt.Errorf("failed to create domain book: %w", err)
	}

	return domainBook, nil
}

// Delete deletes a book by ID
func (r *BookRepository) DeleteBook(ctx context.Context, id int) error {
	_, err := r.db.NewDelete().Model((*models.Book)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete a book: %w", err)
	}

	return nil
}

func (r *BookRepository) GetBooks(ctx context.Context, categoryIDs []int, limit, offset int) ([]domain.Book, error) {
	var books []models.Book
	query := r.db.NewSelect().Model(&books)
	query.Where("stock > 0")
	if len(categoryIDs) > 0 {
		query.Where("category_id IN (?)", bun.In(categoryIDs))
	}
	if limit > 0 {
		query.Limit(limit)
	}
	if offset > 0 {
		query.Offset(offset)
	}
	query.Order("id")
	err := query.Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get books: %w", err)
	}

	domainBooks := make([]domain.Book, len(books))
	for i, book := range books {
		domainBook, err := bookToDomain(book)
		if err != nil {
			return nil, fmt.Errorf("failed to create domain book: %w", err)
		}

		domainBooks[i] = domainBook
	}

	return domainBooks, nil
}
