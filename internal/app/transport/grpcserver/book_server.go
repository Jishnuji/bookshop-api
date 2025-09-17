package grpcserver

import (
	"context"
	"errors"
	"toptal/internal/app/domain"
	"toptal/internal/app/transport/interfaces"
	bookv1 "toptal/proto/v1/book"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BookServer struct {
	bookv1.UnimplementedBookServiceServer
	bookService interfaces.BookService
}

func NewBookServer(bookService interfaces.BookService) *BookServer {
	return &BookServer{
		bookService: bookService,
	}
}

func (s *BookServer) ListBooks(ctx context.Context, req *bookv1.ListBooksRequest) (*bookv1.ListBooksResponse, error) {
	categoryIds := make([]int, len(req.CategoryId))
	for i, categoryID := range req.CategoryId {
		categoryIds[i] = int(categoryID)
	}

	page := int(req.Page)
	var limit, offset int
	if page > 0 {
		limit = 10
		offset = (page - 1) * limit
	}

	books, err := s.bookService.GetBooks(ctx, categoryIds, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get books: %v", err)
	}

	response := make([]*bookv1.CreateBookResponse, 0, len(books))
	for _, book := range books {
		response = append(response, toGRPCBookResponse(book))
	}

	return &bookv1.ListBooksResponse{
		Books: response,
	}, nil
}

func (s *BookServer) GetBook(ctx context.Context, req *bookv1.GetBookRequest) (*bookv1.GetBookResponse, error) {
	if req.Id <= 0 {
		//return convertDomainError(err) TODO: add this
		return nil, status.Errorf(codes.InvalidArgument, "id is required and must be greater than 0")

	}

	book, err := s.bookService.GetBook(ctx, int(req.Id))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "book not found: %v", err)
		}
		return nil, toSlugError(err)
	}

	return &bookv1.GetBookResponse{
		Id:   int64(book.ID()),
		Book: toGRPCBookData(book),
	}, nil
}

func (s *BookServer) CreateBook(ctx context.Context, req *bookv1.CreateBookRequest) (*bookv1.CreateBookResponse, error) {
	domainBook, err := toDomainBook(req.Book)
	if err != nil {
		return nil, toSlugError(err)
	}

	book, err := s.bookService.CreateBook(ctx, domainBook)
	if err != nil {
		return nil, toSlugError(err)
	}

	return toGRPCBookResponse(book), nil
}

func (s *BookServer) UpdateBook(ctx context.Context, req *bookv1.UpdateBookRequest) (*bookv1.UpdateBookResponse, error) {
	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required and must be greater than 0")

	}

	_, err := s.bookService.GetBook(ctx, int(req.Id))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "book not found: %v", err)
		}
		return nil, toSlugError(err)
	}

	domainBook, err := domain.NewBook(domain.NewBookData{
		ID:         int(req.Id),
		Title:      req.Book.Title,
		Year:       int(req.Book.Year),
		Author:     req.Book.Author,
		Price:      int(req.Book.Price),
		Stock:      int(req.Book.Stock),
		CategoryID: int(req.Book.CategoryId),
	})
	if err != nil {
		return nil, toSlugError(err)
	}

	book, err := s.bookService.UpdateBook(ctx, domainBook)
	if err != nil {
		return nil, toSlugError(err)
	}
	return &bookv1.UpdateBookResponse{
		Id:   int64(book.ID()),
		Book: toGRPCBookData(book),
	}, nil
}

func (s *BookServer) DeleteBook(ctx context.Context, req *bookv1.DeleteBookRequest) (*bookv1.DeleteBookResponse, error) {
	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required and must be greater than 0")
	}

	_, err := s.bookService.GetBook(ctx, int(req.Id))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "book not found: %v", err)
		}
		return nil, toSlugError(err)
	}

	err = s.bookService.DeleteBook(ctx, int(req.Id))
	if err != nil {
		return nil, toSlugError(err)
	}
	return &bookv1.DeleteBookResponse{
		Success: true,
	}, nil
}
