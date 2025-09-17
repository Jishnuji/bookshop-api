package grpcserver

import (
	"context"
	"errors"
	"toptal/internal/app/domain"
	"toptal/internal/app/transport/interfaces"
	categoryv1 "toptal/proto/v1/category"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CategoryServer struct {
	categoryv1.UnimplementedCategoryServiceServer
	categoryService interfaces.CategoryService
}

func NewCategoryServer(categoryService interfaces.CategoryService) *CategoryServer {
	return &CategoryServer{
		categoryService: categoryService,
	}
}

func (s *CategoryServer) ListCategories(ctx context.Context, req *categoryv1.ListCategoriesRequest) (*categoryv1.ListCategoriesResponse, error) {
	categories, err := s.categoryService.GetCategories(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get categories: %v", err)
	}

	response := make([]*categoryv1.CreateCategoryResponse, 0, len(categories))
	for _, category := range categories {
		response = append(response, toGRPCCategoryResponse(category))
	}

	return &categoryv1.ListCategoriesResponse{
		Categories: response,
	}, nil
}

func (s *CategoryServer) GetCategory(ctx context.Context, req *categoryv1.GetCategoryRequest) (*categoryv1.GetCategoryResponse, error) {
	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required and must be greater than 0")
	}

	category, err := s.categoryService.GetCategory(ctx, int(req.Id))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "category not found: %v", err)
		}
		return nil, toSlugError(err)
	}

	return &categoryv1.GetCategoryResponse{
		Id:       int64(category.ID()),
		Category: toGRPCCategoryData(category),
	}, nil
}

func (s *CategoryServer) CreateCategory(ctx context.Context, req *categoryv1.CreateCategoryRequest) (*categoryv1.CreateCategoryResponse, error) {
	domainCategory, err := toDomainCategory(req.Category)
	if err != nil {
		return nil, toSlugError(err)
	}

	category, err := s.categoryService.CreateCategory(ctx, domainCategory)
	if err != nil {
		return nil, toSlugError(err)
	}

	return toGRPCCategoryResponse(category), nil
}

func (s *CategoryServer) UpdateCategory(ctx context.Context, req *categoryv1.UpdateCategoryRequest) (*categoryv1.UpdateCategoryResponse, error) {
	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required and must be greater than 0")
	}

	// Verify the category exists
	_, err := s.categoryService.GetCategory(ctx, int(req.Id))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "category not found: %v", err)
		}
		return nil, toSlugError(err)
	}

	domainCategory, err := domain.NewCategory(domain.NewCategoryData{
		ID:   int(req.Id),
		Name: req.Category.Name,
	})
	if err != nil {
		return nil, toSlugError(err)
	}

	category, err := s.categoryService.UpdateCategory(ctx, domainCategory)
	if err != nil {
		return nil, toSlugError(err)
	}

	return &categoryv1.UpdateCategoryResponse{
		Id:       int64(category.ID()),
		Category: toGRPCCategoryData(category),
	}, nil
}

func (s *CategoryServer) DeleteCategory(ctx context.Context, req *categoryv1.DeleteCategoryRequest) (*categoryv1.DeleteCategoryResponse, error) {
	if req.Id <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id is required and must be greater than 0")
	}

	// Verify the category exists
	_, err := s.categoryService.GetCategory(ctx, int(req.Id))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "category not found: %v", err)
		}
		return nil, toSlugError(err)
	}

	err = s.categoryService.DeleteCategory(ctx, int(req.Id))
	if err != nil {
		return nil, toSlugError(err)
	}

	return &categoryv1.DeleteCategoryResponse{
		Success: true,
	}, nil
}
