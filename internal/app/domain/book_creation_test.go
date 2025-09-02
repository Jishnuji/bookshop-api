package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test successful book creation - main scenario
func TestNewBook_ValidData_Success(t *testing.T) {
	// Arrange
	bookData := NewBookData{
		ID:         1,
		Title:      "Clean Architecture",
		Author:     "Robert Martin",
		Year:       2017,
		Price:      2999, // price in cents
		Stock:      15,
		CategoryID: 2,
	}

	// Act
	book, err := NewBook(bookData)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, bookData.ID, book.ID())
	assert.Equal(t, bookData.Title, book.Title())
	assert.Equal(t, bookData.Author, book.Author())
	assert.Equal(t, bookData.Year, book.Year())
	assert.Equal(t, bookData.Price, book.Price())
	assert.Equal(t, bookData.Stock, book.Stock())
	assert.Equal(t, bookData.CategoryID, book.CategoryID())
}

// Test business rule: Title is required
func TestNewBook_EmptyTitle_ReturnsRequiredError(t *testing.T) {
	// Arrange
	bookData := NewBookData{
		ID:         1,
		Title:      "", // business rule violation
		Author:     "Valid Author",
		Year:       2024,
		Price:      1500,
		Stock:      10,
		CategoryID: 1,
	}

	// Act
	book, err := NewBook(bookData)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRequired)
	assert.Contains(t, err.Error(), "title")
	assert.Equal(t, Book{}, book)
}

// Test business rule: Author is required
func TestNewBook_EmptyAuthor_ReturnsRequiredError(t *testing.T) {
	// Arrange
	bookData := NewBookData{
		ID:         1,
		Title:      "Valid Title",
		Author:     "", // business rule violation
		Year:       2024,
		Price:      1500,
		Stock:      10,
		CategoryID: 1,
	}

	// Act
	book, err := NewBook(bookData)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRequired)
	assert.Contains(t, err.Error(), "author")
	assert.Equal(t, Book{}, book)
}

// Test business rule: Year must be positive
func TestNewBook_InvalidYear_ReturnsNegativeError(t *testing.T) {
	testCases := []struct {
		name string
		year int
	}{
		{"Zero year", 0},
		{"Negative year", -1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			bookData := NewBookData{
				ID:         1,
				Title:      "Valid Title",
				Author:     "Valid Author",
				Year:       tc.year, // business rule violation
				Price:      1500,
				Stock:      10,
				CategoryID: 1,
			}

			// Act
			book, err := NewBook(bookData)

			// Assert
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrNegative)
			assert.Contains(t, err.Error(), "year")
			assert.Equal(t, Book{}, book)
		})
	}
}

// Test business rule: Price must be positive
func TestNewBook_InvalidPrice_ReturnsNegativeError(t *testing.T) {
	testCases := []struct {
		name  string
		price int
	}{
		{"Zero price", 0},
		{"Negative price", -100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			bookData := NewBookData{
				ID:         1,
				Title:      "Valid Title",
				Author:     "Valid Author",
				Year:       2024,
				Price:      tc.price, // business rule violation
				Stock:      10,
				CategoryID: 1,
			}

			// Act
			book, err := NewBook(bookData)

			// Assert
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrNegative)
			assert.Contains(t, err.Error(), "price")
			assert.Equal(t, Book{}, book)
		})
	}
}

// Test business rule: Stock cannot be negative (zero is allowed)
func TestNewBook_NegativeStock_ReturnsNegativeError(t *testing.T) {
	// Arrange
	bookData := NewBookData{
		ID:         1,
		Title:      "Valid Title",
		Author:     "Valid Author",
		Year:       2024,
		Price:      1500,
		Stock:      -1, // business rule violation
		CategoryID: 1,
	}

	// Act
	book, err := NewBook(bookData)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNegative)
	assert.Contains(t, err.Error(), "stock")
	assert.Equal(t, Book{}, book)
}

// Test business rule: Zero stock is allowed (item is out of stock)
func TestNewBook_ZeroStock_Success(t *testing.T) {
	// Arrange
	bookData := NewBookData{
		ID:         1,
		Title:      "Out of Stock Book",
		Author:     "Valid Author",
		Year:       2024,
		Price:      1500,
		Stock:      0, // allowed value
		CategoryID: 1,
	}

	// Act
	book, err := NewBook(bookData)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0, book.Stock())
}

// Test business rule: CategoryID is required
func TestNewBook_ZeroCategoryID_ReturnsRequiredError(t *testing.T) {
	// Arrange
	bookData := NewBookData{
		ID:         1,
		Title:      "Valid Title",
		Author:     "Valid Author",
		Year:       2024,
		Price:      1500,
		Stock:      10,
		CategoryID: 0, // business rule violation
	}

	// Act
	book, err := NewBook(bookData)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRequired)
	assert.Contains(t, err.Error(), "category_id")
	assert.Equal(t, Book{}, book)
}

// Test business rule: ID can be zero for new books
func TestNewBook_ZeroID_Success(t *testing.T) {
	// Arrange - new book without ID (will be assigned by database)
	bookData := NewBookData{
		ID:         0, // допустимо для новых записей
		Title:      "New Book",
		Author:     "New Author",
		Year:       2024,
		Price:      1000,
		Stock:      5,
		CategoryID: 1,
	}

	// Act
	book, err := NewBook(bookData)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0, book.ID())
	assert.Equal(t, "New Book", book.Title())
}

// Test business rules validation order
func TestNewBook_ValidationOrder_ReturnsFirstError(t *testing.T) {
	// Arrange - all rules are violated, but should return the first error
	bookData := NewBookData{
		ID:         1,
		Title:      "", // первая ошибка по порядку валидации
		Author:     "",
		Year:       0,
		Price:      0,
		Stock:      -1,
		CategoryID: 0,
	}

	// Act
	book, err := NewBook(bookData)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRequired)
	assert.Contains(t, err.Error(), "title") // первая ошибка
	assert.Equal(t, Book{}, book)
}

// Test correct work of all getters
func TestBook_Getters_ReturnCorrectValues(t *testing.T) {
	// Arrange
	expectedData := NewBookData{
		ID:         42,
		Title:      "Domain-Driven Design",
		Author:     "Eric Evans",
		Year:       2003,
		Price:      3500,
		Stock:      8,
		CategoryID: 3,
	}

	book, err := NewBook(expectedData)
	require.NoError(t, err)

	// Act & Assert - verify encapsulation through getters
	assert.Equal(t, expectedData.ID, book.ID())
	assert.Equal(t, expectedData.Title, book.Title())
	assert.Equal(t, expectedData.Author, book.Author())
	assert.Equal(t, expectedData.Year, book.Year())
	assert.Equal(t, expectedData.Price, book.Price())
	assert.Equal(t, expectedData.Stock, book.Stock())
	assert.Equal(t, expectedData.CategoryID, book.CategoryID())
}
