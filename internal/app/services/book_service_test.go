package services

import (
	"context"
	"errors"
	"testing"
	"toptal/internal/app/domain"
	"toptal/internal/app/services/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookService_GetBook_Success(t *testing.T) {
	// Arrange
	mockRepo := mocks.NewMockBookRepository(t)
	service := NewBookService(mockRepo)
	ctx := context.Background()
	bookID := 1

	expectedBook, _ := domain.NewBook(domain.NewBookData{
		ID:         1,
		Title:      "The Go Programming Language",
		Author:     "Alan Donovan",
		Year:       2015,
		Price:      4999,
		Stock:      5,
		CategoryID: 1,
	})

	mockRepo.EXPECT().
		GetBook(ctx, bookID).
		Return(expectedBook, nil).
		Once()

	// Act
	result, err := service.GetBook(ctx, bookID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedBook, result)
}

func TestBookService_GetBook_InvalidID(t *testing.T) {
	// Arrange
	mockRepo := mocks.NewMockBookRepository(t)
	service := NewBookService(mockRepo)
	ctx := context.Background()

	// Act
	result, err := service.GetBook(ctx, 0)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrRequired)
	assert.Equal(t, domain.Book{}, result)
}

func TestBookService_GetBook_NotFound(t *testing.T) {
	// Arrange
	mockRepo := mocks.NewMockBookRepository(t)
	service := NewBookService(mockRepo)
	ctx := context.Background()
	bookID := 999

	mockRepo.EXPECT().
		GetBook(ctx, bookID).
		Return(domain.Book{}, domain.ErrNotFound).
		Once()

	// Act
	result, err := service.GetBook(ctx, bookID)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Equal(t, domain.Book{}, result)
}

func TestBookService_GetBooks_Success(t *testing.T) {
	// Arrange
	mockRepo := mocks.NewMockBookRepository(t)
	service := NewBookService(mockRepo)
	ctx := context.Background()
	categoryIDs := []int{1, 2}
	limit := 10
	offset := 0

	expectedBooks := []domain.Book{
		// создайте тестовые книги через domain.NewBook()
	}

	mockRepo.EXPECT().
		GetBooks(ctx, categoryIDs, limit, offset).
		Return(expectedBooks, nil).
		Once()

	// Act
	result, err := service.GetBooks(ctx, categoryIDs, limit, offset)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedBooks, result)
	assert.Len(t, result, len(expectedBooks))
}

func TestBookService_GetBooks_EmptyResult(t *testing.T) {
	// Arrange
	mockRepo := mocks.NewMockBookRepository(t)
	service := NewBookService(mockRepo)
	ctx := context.Background()

	mockRepo.EXPECT().
		GetBooks(ctx, []int{}, 0, 0).
		Return([]domain.Book{}, nil).
		Once()

	// Act
	result, err := service.GetBooks(ctx, []int{}, 0, 0)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestBookService_GetBooks_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := mocks.NewMockBookRepository(t)
	service := NewBookService(mockRepo)
	ctx := context.Background()
	expectedError := errors.New("database connection failed")

	mockRepo.EXPECT().
		GetBooks(ctx, []int{}, 0, 0).
		Return(nil, expectedError).
		Once()

	// Act
	result, err := service.GetBooks(ctx, []int{}, 0, 0)

	// Assert
	require.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestBookService_CreateBook_Success(t *testing.T) {
	// Arrange
	mockRepo := mocks.NewMockBookRepository(t)
	service := NewBookService(mockRepo)
	ctx := context.Background()

	inputBook := domain.Book{}    // создайте через domain.NewBook()
	expectedBook := domain.Book{} // с ID после создания

	mockRepo.EXPECT().
		CreateBook(ctx, inputBook).
		Return(expectedBook, nil).
		Once()

	// Act
	result, err := service.CreateBook(ctx, inputBook)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedBook, result)
}

func TestBookService_UpdateBook_Success(t *testing.T) {
	// Arrange
	mockRepo := mocks.NewMockBookRepository(t)
	service := NewBookService(mockRepo)
	ctx := context.Background()

	inputBook := domain.Book{}    // создайте с ID
	expectedBook := domain.Book{} // обновленная версия

	mockRepo.EXPECT().
		UpdateBook(ctx, inputBook).
		Return(expectedBook, nil).
		Once()

	// Act
	result, err := service.UpdateBook(ctx, inputBook)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedBook, result)
}

func TestBookService_DeleteBook_Success(t *testing.T) {
	// Arrange
	mockRepo := mocks.NewMockBookRepository(t)
	service := NewBookService(mockRepo)
	ctx := context.Background()
	bookID := 1

	mockRepo.EXPECT().
		DeleteBook(ctx, bookID).
		Return(nil).
		Once()

	// Act
	err := service.DeleteBook(ctx, bookID)

	// Assert
	require.NoError(t, err)
}

func TestBookService_DeleteBook_InvalidID(t *testing.T) {
	// Arrange
	mockRepo := mocks.NewMockBookRepository(t)
	service := NewBookService(mockRepo)
	ctx := context.Background()

	// Act
	err := service.DeleteBook(ctx, 0)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrRequired)
}

// ... existing code ...

func TestBookService_CreateBook_InvalidCategory(t *testing.T) {
	// Arrange
	mockRepo := mocks.NewMockBookRepository(t)
	service := NewBookService(mockRepo)
	ctx := context.Background()

	inputBook, _ := domain.NewBook(domain.NewBookData{
		ID:         0,
		Title:      "Test Book",
		Author:     "Test Author",
		Year:       2024,
		Price:      1500,
		Stock:      10,
		CategoryID: 1999,
	})

	expectedError := errors.New("category not found")

	mockRepo.EXPECT().
		CreateBook(ctx, inputBook).
		Return(domain.Book{}, expectedError).
		Once()

	// Act
	result, err := service.CreateBook(ctx, inputBook)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "category not found")
	assert.Equal(t, domain.Book{}, result)
}

func TestBookService_GetBooks_InvalidCategory(t *testing.T) {
	// Arrange
	mockRepo := mocks.NewMockBookRepository(t)
	service := NewBookService(mockRepo)
	ctx := context.Background()

	expectedError := errors.New("category not found")

	mockRepo.EXPECT().
		GetBooks(ctx, []int{1999}, 10, 0).
		Return(nil, expectedError).
		Once()

	// Act
	result, err := service.GetBooks(ctx, []int{1999}, 10, 0)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "category not found")
	assert.Nil(t, result)
}
