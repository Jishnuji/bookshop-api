package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"toptal/internal/app/common/server"
	"toptal/internal/app/domain"
	"toptal/internal/app/transport/models"

	"github.com/go-chi/chi/v5"
)

func (s HttpServer) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := s.categoryService.GetCategories(r.Context())
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	response := make([]models.CategoryResponse, 0, len(categories))
	for _, category := range categories {
		response = append(response, toResponseCategory(category))
	}

	server.RespondOK(response, w, r)
}

// GetCategory returns a category by ID
func (s HttpServer) GetCategory(w http.ResponseWriter, r *http.Request) {
	categoryIDParam := chi.URLParam(r, "category_id")
	categoryID, err := strconv.Atoi(categoryIDParam)
	if err != nil {
		server.BadRequest("invalid-category-id", err, w, r)
		return
	}
	category, err := s.categoryService.GetCategory(r.Context(), categoryID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			server.NotFound("category-not-found", err, w, r)
			return
		}
		server.RespondWithError(err, w, r)
		return
	}

	response := toResponseCategory(category)

	server.RespondOK(response, w, r)
}

// CreateCategory creates a new category
func (s HttpServer) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var categoryRequest models.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&categoryRequest); err != nil {
		server.BadRequest("invalid-json", err, w, r)
		return
	}

	category, err := domain.NewCategory(domain.NewCategoryData{
		Name: categoryRequest.Name,
	})
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	insertedCategory, err := s.categoryService.CreateCategory(r.Context(), category)
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	response := toResponseCategory(insertedCategory)

	server.RespondOK(response, w, r)
}

func (s HttpServer) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	categoryIDParam := chi.URLParam(r, "category_id")
	categoryID, err := strconv.Atoi(categoryIDParam)
	if err != nil {
		server.BadRequest("invalid-category-id", err, w, r)
		return
	}

	var categoryRequest models.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&categoryRequest); err != nil {
		server.BadRequest("invalid-json", err, w, r)
		return
	}

	_, err = s.categoryService.GetCategory(r.Context(), categoryID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			server.NotFound("category-not-found", err, w, r)
			return
		}
		server.RespondWithError(err, w, r)
		return
	}

	category, err := domain.NewCategory(domain.NewCategoryData{
		ID:   categoryID,
		Name: categoryRequest.Name,
	})
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	updatedCategory, err := s.categoryService.UpdateCategory(r.Context(), category)
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	response := toResponseCategory(updatedCategory)

	server.RespondOK(response, w, r)
}

func (s HttpServer) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	categoryIDParam := chi.URLParam(r, "category_id")
	categoryID, err := strconv.Atoi(categoryIDParam)
	if err != nil {
		server.BadRequest("invalid-category-id", err, w, r)
		return
	}

	_, err = s.categoryService.GetCategory(r.Context(), categoryID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			server.NotFound("category-not-found", err, w, r)
			return
		}
		server.RespondWithError(err, w, r)
		return
	}

	err = s.categoryService.DeleteCategory(r.Context(), categoryID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			server.NotFound("category-not-found", err, w, r)
			return
		}
		server.RespondWithError(err, w, r)
		return
	}

	server.RespondOK(map[string]bool{"deleted": true}, w, r)
}
