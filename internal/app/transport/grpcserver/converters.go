package grpcserver

import (
	"errors"
	"toptal/internal/app/common/slugerrors"
	"toptal/internal/app/domain"
	bookv1 "toptal/proto/v1/book"
	cartv1 "toptal/proto/v1/cart"
	categoryv1 "toptal/proto/v1/category"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCBookResponse(book domain.Book) *bookv1.CreateBookResponse {
	return &bookv1.CreateBookResponse{
		Id:   int64(book.ID()),
		Book: toGRPCBookData(book),
	}
}

func toGRPCBookData(book domain.Book) *bookv1.BookData {
	return &bookv1.BookData{
		Title:      book.Title(),
		Year:       int32(book.Year()),
		Author:     book.Author(),
		Price:      int32(book.Price()),
		Stock:      int32(book.Stock()),
		CategoryId: int32(book.CategoryID()),
	}
}

func toDomainBook(bookRequest *bookv1.BookData) (domain.Book, error) {
	return domain.NewBook(domain.NewBookData{
		Title:      bookRequest.Title,
		Year:       int(bookRequest.Year),
		Author:     bookRequest.Author,
		Price:      int(bookRequest.Price),
		Stock:      int(bookRequest.Stock),
		CategoryID: int(bookRequest.CategoryId),
	})
}

func toGRPCCategoryResponse(category domain.Category) *categoryv1.CreateCategoryResponse {
	return &categoryv1.CreateCategoryResponse{
		Id:       int64(category.ID()),
		Category: toGRPCCategoryData(category),
	}
}

func toGRPCCategoryData(category domain.Category) *categoryv1.CategoryData {
	return &categoryv1.CategoryData{
		Name: category.Name(),
	}
}

func toDomainCategory(categoryRequest *categoryv1.CategoryData) (domain.Category, error) {
	return domain.NewCategory(domain.NewCategoryData{
		Name: categoryRequest.Name,
	})
}

// Cart converters
func toGRPCCartData(cart domain.Cart) *cartv1.CartData {
	bookIDs := make([]int64, len(cart.BookIDs()))
	for i, id := range cart.BookIDs() {
		bookIDs[i] = int64(id)
	}

	return &cartv1.CartData{
		BookIds: bookIDs,
	}
}

func toDomainCartFromGRPC(userID int, cartData *cartv1.CartData) (domain.Cart, error) {
	bookIDs := make([]int, len(cartData.BookIds))
	for i, id := range cartData.BookIds {
		bookIDs[i] = int(id)
	}

	return domain.NewCart(domain.NewCartData{
		UserID:  userID,
		BookIDs: bookIDs,
	})
}

// Error converters
func toSlugError(err error) error {
	var slugError slugerrors.SlugError
	if !errors.As(err, &slugError) {
		return status.Error(codes.Internal, "internal server error")
	}

	switch slugError.ErrorType() {
	case slugerrors.ErrorTypeAuthorization:
		return status.Error(codes.Unauthenticated, slugError.Error())
	case slugerrors.ErrorTypeBadRequest:
		return status.Error(codes.InvalidArgument, slugError.Error())
	case slugerrors.ErrorTypeNotFound:
		return status.Error(codes.NotFound, slugError.Error())
	default:
		return status.Error(codes.Internal, slugError.Error())
	}
}

func convertDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return status.Errorf(codes.NotFound, "book not found")
	//case errors.Is(err, domain.ErrInvalidInput):
	//	return status.Errorf(codes.InvalidArgument, "invalid input: %v", err)
	//case errors.Is(err, domain.ErrUnauthorized):
	//	return status.Errorf(codes.Unauthenticated, "unauthorized")
	//case errors.Is(err, domain.ErrPermissionDenied):
	//	return status.Errorf(codes.PermissionDenied, "permission denied")
	default:
		return status.Errorf(codes.Internal, "internal error: %v", err)
	}
}
