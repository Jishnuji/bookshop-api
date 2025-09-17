package grpcserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"toptal/internal/app/domain"
	"toptal/internal/app/services"
	"toptal/internal/app/services/mocks"
	"toptal/internal/app/transport/grpcserver"
	"toptal/internal/app/transport/models"
	bookv1 "toptal/proto/v1/book"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func TestCreateBook_InvalidYear_ThroughGateway(t *testing.T) {
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
			expectedStatus:    http.StatusInternalServerError, // gRPC Gateway returns 500
			expectedErrorText: "",                             // gateway returns generic "internal server error"
			description:       "Zero year should trigger validation error through gRPC Gateway",
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
			expectedStatus:    http.StatusInternalServerError, // gRPC Gateway returns 500
			expectedErrorText: "",                             // gateway returns generic "internal server error"
			description:       "Negative year should trigger validation error through gRPC Gateway",
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
			description:       "Valid year should create book successfully through gRPC Gateway",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange - create test gRPC Gateway with mock dependencies
			gatewayServer := setupGatewayTestServer(t)
			defer gatewayServer.Close()

			wrapped := map[string]interface{}{"book": tc.bookRequest}
			requestBody, err := json.Marshal(wrapped)
			require.NoError(t, err)

			// Act - execute HTTP request through gRPC Gateway
			resp, err := http.Post(gatewayServer.URL+"/v1/book", "application/json", bytes.NewReader(requestBody))
			require.NoError(t, err)
			defer resp.Body.Close()

			// Assert - check result
			assert.Equal(t, tc.expectedStatus, resp.StatusCode, "Unexpected HTTP status")

			var responseBody map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&responseBody)
			require.NoError(t, err)

			responseBytes, _ := json.Marshal(responseBody)
			t.Logf("✓ HTTP Status: %d", resp.StatusCode)
			t.Logf("✓ Response: %s", string(responseBytes))

			if tc.expectedStatus == http.StatusOK {
				// Success case - nested response: { "id": ..., "book": { ... } }
				assert.NotEmpty(t, responseBody["id"])

				bookObj, ok := responseBody["book"].(map[string]interface{})
				require.True(t, ok, "response.book must be an object")

				assert.Equal(t, tc.bookRequest.Title, bookObj["title"])
				assert.Equal(t, tc.bookRequest.Author, bookObj["author"])
				assert.Equal(t, float64(tc.bookRequest.Year), bookObj["year"])

				t.Logf("✓ Book created through Gateway: %s by %s (%v)",
					bookObj["title"], bookObj["author"], bookObj["year"])
			} else {
				// Error case - gRPC Gateway error format
				assert.Contains(t, responseBody, "message")
				assert.Contains(t, responseBody, "code")
				assert.Equal(t, float64(13), responseBody["code"])
				assert.Equal(t, "internal server error", responseBody["message"])

				// Optional: field-specific text is not present in gateway error
				if tc.expectedErrorText != "" && responseBody["message"] != nil {
					if errorText, ok := responseBody["message"].(string); ok && tc.expectedErrorText != "" {
						assert.Contains(t, errorText, tc.expectedErrorText,
							"Error should contain information about field '%s'", tc.expectedErrorText)
						t.Logf("✓ Validation caught error through Gateway: %s", errorText)
					}
				}
			}

			t.Logf("✓ %s: %s", tc.name, tc.description)
		})
	}
}

// setupGatewayTestServer creates gRPC Gateway test server with mock dependencies
func setupGatewayTestServer(t *testing.T) *httptest.Server {
	// Create mock repository for BookService
	mockBookRepo := &mocks.MockBookRepository{}

	// Configure mock for successful valid book creation
	createdBook := createValidBook(t)
	mockBookRepo.On("CreateBook", mock.Anything, mock.MatchedBy(func(book domain.Book) bool {
		return book.Year() > 0 // success only for a valid year
	})).Return(createdBook, nil)

	// Create real BookService with mock repository
	bookService := services.NewBookService(mockBookRepo)

	// Setup in-process gRPC server
	lis := bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()

	// Register gRPC services
	bookServer := grpcserver.NewBookServer(bookService)
	bookv1.RegisterBookServiceServer(grpcServer, bookServer)

	// Start gRPC server in background
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("gRPC server error: %v", err)
		}
	}()

	// Setup gRPC Gateway
	ctx := context.Background()
	mux := runtime.NewServeMux()

	// Create connection to in-process gRPC server
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	// Register gRPC Gateway handlers
	err = bookv1.RegisterBookServiceHandler(ctx, mux, conn)
	require.NoError(t, err)

	// Create HTTP test server with gRPC Gateway
	httpServer := httptest.NewServer(mux)

	// Cleanup function
	t.Cleanup(func() {
		conn.Close()
		grpcServer.Stop()
		httpServer.Close()
	})

	return httpServer
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

// Test demonstrating architectural findings in gRPC Gateway
func TestCreateBook_ArchitecturalFindings_ThroughGateway(t *testing.T) {
	t.Run("Domain validation errors return HTTP 500 through gRPC Gateway", func(t *testing.T) {
		// Arrange
		gatewayServer := setupGatewayTestServer(t)
		defer gatewayServer.Close()

		// Request violating business rule
		invalidRequest := models.BookRequest{
			Title:      "Test Book",
			Author:     "Test Author",
			Year:       0, // violation: year must be > 0
			Price:      1000,
			CategoryID: 1,
		}

		wrapped := map[string]interface{}{"book": invalidRequest}
		requestBody, err := json.Marshal(wrapped)
		require.NoError(t, err)

		// Act
		resp, err := http.Post(gatewayServer.URL+"/v1/book", "application/json", bytes.NewReader(requestBody))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Assert - document current behavior
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode,
			"ARCHITECTURAL ISSUE: domain validation errors return 500 through gRPC Gateway")

		var responseBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseBody)
		require.NoError(t, err)

		t.Log("✓ ARCHITECTURAL ISSUE FOUND (gRPC Gateway):")
		t.Log("  - Domain errors (ErrNegative, ErrRequired) are not handled as proper gRPC status codes")
		t.Log("  - gRPC Gateway converts Internal gRPC errors to HTTP 500")
		t.Log("  - Client receives 500 Internal Server Error instead of 400 Bad Request")
		t.Log("  - gRPC error format: {\"code\": 13, \"message\": \"...\", \"details\": [...]}")
		t.Logf("  - Actual response: %+v", responseBody)
		t.Log("")
		t.Log("✓ RECOMMENDATION: use proper gRPC status codes (InvalidArgument) for domain errors")
	})

	t.Run("gRPC Gateway error format differs from direct HTTP", func(t *testing.T) {
		// This test documents the difference in error response format
		// between direct HTTP handlers and gRPC Gateway

		gatewayServer := setupGatewayTestServer(t)
		defer gatewayServer.Close()

		invalidRequest := models.BookRequest{
			Title:      "Test Book",
			Author:     "Test Author",
			Year:       -1,
			Price:      1000,
			CategoryID: 1,
		}

		wrapped := map[string]interface{}{"book": invalidRequest}
		requestBody, _ := json.Marshal(wrapped)
		resp, _ := http.Post(gatewayServer.URL+"/v1/book", "application/json", bytes.NewReader(requestBody))
		defer resp.Body.Close()

		var gatewayError map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&gatewayError)

		t.Log("✓ gRPC Gateway error format:")
		t.Logf("  - Uses 'message' field instead of 'slug'/'error'")
		t.Logf("  - May include 'code' field with gRPC status code")
		t.Logf("  - Structure: %+v", gatewayError)
		t.Log("")
		t.Log("✓ NOTE: Error handling should be unified across HTTP and gRPC Gateway")
	})
}
