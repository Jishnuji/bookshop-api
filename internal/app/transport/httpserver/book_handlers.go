package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	auth "toptal/internal/app/common/auth"
	"toptal/internal/app/common/server"
	"toptal/internal/app/domain"
	"toptal/internal/app/transport/models"

	"github.com/go-chi/chi/v5"
)

func (s HttpServer) GetBooks(w http.ResponseWriter, r *http.Request) {
	// filter by category IDs
	queryCategoryIDs := r.URL.Query()["category_id"]
	var categoryIDs []int
	for _, id := range queryCategoryIDs {
		categoryID, err := strconv.Atoi(id)
		if err != nil {
			server.BadRequest("invalid-category-id", err, w, r)
			return
		}
		categoryIDs = append(categoryIDs, categoryID)
	}
	// page
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}
	var limit, offset int
	if page > 0 {
		limit = 10
		offset = (page - 1) * limit
	}

	books, err := s.bookService.GetBooks(r.Context(), categoryIDs, limit, offset)
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	response := make([]models.BookResponse, 0, len(books))
	for _, book := range books {
		response = append(response, auth.ToResponseBook(book))
	}

	server.RespondOK(response, w, r)
}

func (s HttpServer) GetBook(w http.ResponseWriter, r *http.Request) {
	bookIDParam := chi.URLParam(r, "book_id")
	bookID, err := strconv.Atoi(bookIDParam)
	if err != nil {
		server.BadRequest("invalid-book-id", err, w, r)
		return
	}
	book, err := s.bookService.GetBook(r.Context(), bookID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			server.NotFound("book-not-found", err, w, r)
			return
		}
		server.RespondWithError(err, w, r)
		return
	}
	response := auth.ToResponseBook(book)

	server.RespondOK(response, w, r)
}

// CreateBook creates a new book
func (s HttpServer) CreateBook(w http.ResponseWriter, r *http.Request) {
	var bookRequest models.BookRequest
	if err := json.NewDecoder(r.Body).Decode(&bookRequest); err != nil {
		server.BadRequest("invalid-json", err, w, r)
		return
	}

	book, err := auth.ToDomainBook(bookRequest)
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	insertedBook, err := s.bookService.CreateBook(r.Context(), book)
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	response := auth.ToResponseBook(insertedBook)

	server.RespondOK(response, w, r)
}

// UpdateBook updates a book by ID
func (s HttpServer) UpdateBook(w http.ResponseWriter, r *http.Request) {
	bookIDParam := chi.URLParam(r, "book_id")
	bookID, err := strconv.Atoi(bookIDParam)
	if err != nil {
		server.BadRequest("invalid-book-id", err, w, r)
		return
	}

	var bookRequest models.BookRequest
	if err := json.NewDecoder(r.Body).Decode(&bookRequest); err != nil {
		server.BadRequest("invalid-json", err, w, r)
		return
	}

	_, err = s.bookService.GetBook(r.Context(), bookID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			server.NotFound("book-not-found", err, w, r)
			return
		}
		server.RespondWithError(err, w, r)
		return
	}

	book, err := domain.NewBook(domain.NewBookData{
		ID:         bookID,
		Title:      bookRequest.Title,
		Year:       bookRequest.Year,
		Author:     bookRequest.Author,
		Price:      bookRequest.Price,
		CategoryID: bookRequest.CategoryID,
	})
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	updatedBook, err := s.bookService.UpdateBook(r.Context(), book)
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	response := auth.ToResponseBook(updatedBook)

	server.RespondOK(response, w, r)
}

// DeleteBook deletes a book by ID
func (s HttpServer) DeleteBook(w http.ResponseWriter, r *http.Request) {
	bookIDParam := chi.URLParam(r, "book_id")
	bookID, err := strconv.Atoi(bookIDParam)
	if err != nil {
		server.BadRequest("invalid-book-id", err, w, r)
		return
	}

	_, err = s.bookService.GetBook(r.Context(), bookID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			server.NotFound("book-not-found", err, w, r)
			return
		}
		server.RespondWithError(err, w, r)
		return
	}

	err = s.bookService.DeleteBook(r.Context(), bookID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			server.NotFound("book-not-found", err, w, r)
			return
		}
		server.RespondWithError(err, w, r)
		return
	}

	server.RespondOK(map[string]bool{"deleted": true}, w, r)
}
