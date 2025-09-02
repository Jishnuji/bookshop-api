package httpserver_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"toptal/internal/app/domain"
	"toptal/internal/app/services"
	"toptal/internal/app/services/mocks"
	"toptal/internal/app/transport/httpserver"
	"toptal/internal/app/transport/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateBook_InvalidYear_Integration(t *testing.T) {
	// Enable DEBUG_ERRORS for error visibility
	oldDebug := os.Getenv("DEBUG_ERRORS")
	os.Setenv("DEBUG_ERRORS", "1")
	defer os.Setenv("DEBUG_ERRORS", oldDebug)

	testCases := []struct {
		name              string
		bookRequest       models.BookRequest
		expectedStatus    int
		expectedErrorText string
		description       string
	}{
		{
			name: "Zero year validation",
			bookRequest: models.BookRequest{
				Title:      "Valid Title",
				Author:     "Valid Author",
				Year:       0, // business rule: year must be > 0
				Price:      1500,
				CategoryID: 1,
			},
			expectedStatus:    http.StatusInternalServerError, // system returns 500
			expectedErrorText: "year",                         // in debug mode
			description:       "Zero year should trigger validation error",
		},
		{
			name: "Negative year validation",
			bookRequest: models.BookRequest{
				Title:      "Valid Title",
				Author:     "Valid Author",
				Year:       -1, // business rule violated
				Price:      1500,
				CategoryID: 1,
			},
			expectedStatus:    http.StatusInternalServerError, // system returns 500
			expectedErrorText: "year",
			description:       "Negative year should trigger validation error",
		},
		{
			name: "Valid year success",
			bookRequest: models.BookRequest{
				Title:      "Clean Architecture",
				Author:     "Robert Martin",
				Year:       2017, // valid year
				Price:      2999,
				CategoryID: 1,
			},
			expectedStatus:    http.StatusOK,
			expectedErrorText: "",
			description:       "Valid year should create book successfully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange - create test HTTP server with mock dependencies
			server := setupTestServer(t)

			requestBody, err := json.Marshal(tc.bookRequest)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/books", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Act - execute HTTP request
			w := httptest.NewRecorder()
			server.CreateBook(w, req)

			// Assert - check result
			assert.Equal(t, tc.expectedStatus, w.Code, "Unexpected HTTP status")

			t.Logf("✓ HTTP Status: %d", w.Code)
			t.Logf("✓ Response: %s", w.Body.String())

			if tc.expectedStatus == http.StatusOK {
				// Success case
				var response models.BookResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Equal(t, tc.bookRequest.Title, response.Title)
				assert.Equal(t, tc.bookRequest.Author, response.Author)
				assert.Equal(t, tc.bookRequest.Year, response.Year)

				t.Logf("✓ Book created: %s by %s (%d)", response.Title, response.Author, response.Year)
			} else {
				// Error case
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				require.NoError(t, err)

				// Check that error slug exists
				assert.Contains(t, errorResponse, "slug")

				// In DEBUG mode there should be error description
				if tc.expectedErrorText != "" && errorResponse["error"] != nil {
					if errorObj, ok := errorResponse["error"]; ok {
						if errorText, ok := errorObj.(string); ok && tc.expectedErrorText != "" {
							assert.Contains(t, errorText, tc.expectedErrorText,
								"Error should contain information about field '%s'", tc.expectedErrorText)
							t.Logf("✓ Validation caught error: %s", errorText)
						}
					}
				}
			}

			t.Logf("✓ %s: %s", tc.name, tc.description)
		})
	}
}

// setupTestServer creates HTTP server with mock dependencies
func setupTestServer(t *testing.T) *httpserver.HttpServer {
	// Create mock repository for BookService
	mockBookRepo := &mocks.MockBookRepository{}

	// Configure mock for successful valid book creation
	createdBook := createValidBook(t)
	mockBookRepo.On("CreateBook", mock.Anything, mock.MatchedBy(func(book domain.Book) bool {
		return book.Year() > 0 // success only for a valid year
	})).Return(createdBook, nil)

	// Create real BookService with mock repository
	bookService := services.NewBookService(mockBookRepo)

	// Create HttpServer with real service (other dependencies' nil)
	return httpserver.NewHttpServer(
		nil,         // userService - not needed for this test
		nil,         // authService - not needed for this test
		bookService, // bookService - service being tested
		nil,         // cartService - not needed for this test
		nil,         // categoryService - not needed for this test
	)
}

// createValidBook creates valid domain book for tests
func createValidBook(t *testing.T) domain.Book {
	bookData := domain.NewBookData{
		ID:         1,
		Title:      "Clean Architecture",
		Author:     "Robert Martin",
		Year:       2017,
		Price:      2999,
		Stock:      10,
		CategoryID: 1,
	}

	book, err := domain.NewBook(bookData)
	require.NoError(t, err)
	return book
}

// Test demonstrating architectural findings
func TestCreateBook_ArchitecturalFindings_Integration(t *testing.T) {
	t.Run("Domain validation errors return HTTP 500 instead of 400", func(t *testing.T) {
		// Arrange
		server := setupTestServer(t)

		// Request violating business rule
		invalidRequest := models.BookRequest{
			Title:      "Test Book",
			Author:     "Test Author",
			Year:       0, // violation: year must be > 0
			Price:      1000,
			CategoryID: 1,
		}

		requestBody, err := json.Marshal(invalidRequest)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/books", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")

		// Act
		w := httptest.NewRecorder()
		server.CreateBook(w, req)

		// Assert - document current behavior
		assert.Equal(t, http.StatusInternalServerError, w.Code,
			"ARCHITECTURAL ISSUE: domain validation errors return 500 instead of 400")

		t.Log("✓ ARCHITECTURAL ISSUE FOUND:")
		t.Log("  - Domain errors (ErrNegative, ErrRequired) are not handled as SlugError")
		t.Log("  - HTTP layer does not distinguish between validation and system errors")
		t.Log("  - Client receives 500 Internal Server Error instead of 400 Bad Request")
		t.Log("")
		t.Log("✓ RECOMMENDATION: create wrapper for domain errors in SlugError")
	})
}
