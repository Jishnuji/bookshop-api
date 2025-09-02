package services

import (
	"context"
	"fmt"
	"toptal/internal/app/domain"
)

type BookService struct {
	repo BookRepository
}

func NewBookService(repo BookRepository) *BookService {
	return &BookService{
		repo: repo,
	}
}

func (s BookService) GetBook(ctx context.Context, id int) (domain.Book, error) {
	if id == 0 {
		return domain.Book{}, fmt.Errorf("%w: id", domain.ErrRequired)
	}
	return s.repo.GetBook(ctx, id)
}

func (s BookService) CreateBook(ctx context.Context, book domain.Book) (domain.Book, error) {
	return s.repo.CreateBook(ctx, book)
}

func (s BookService) UpdateBook(ctx context.Context, book domain.Book) (domain.Book, error) {
	return s.repo.UpdateBook(ctx, book)
}

func (s BookService) DeleteBook(ctx context.Context, id int) error {
	if id == 0 {
		return fmt.Errorf("%w: id", domain.ErrRequired)
	}
	return s.repo.DeleteBook(ctx, id)
}

func (s BookService) GetBooks(ctx context.Context, categoryIDs []int, limit, offset int) ([]domain.Book, error) {
	return s.repo.GetBooks(ctx, categoryIDs, limit, offset)
}
