package auth

import (
	"context"
	"strconv"
	"toptal/internal/app/domain"
	"toptal/internal/app/transport/models"

	"google.golang.org/grpc/metadata"
)

func ToResponseBook(book domain.Book) models.BookResponse {
	return models.BookResponse{
		ID:         book.ID(),
		Title:      book.Title(),
		Year:       book.Year(),
		Author:     book.Author(),
		Price:      book.Price(),
		Stock:      book.Stock(),
		CategoryID: book.CategoryID(),
	}
}

func ToResponseCategory(category domain.Category) models.CategoryResponse {
	return models.CategoryResponse{
		ID:   category.ID(),
		Name: category.Name(),
	}
}

func ToDomainBook(bookRequest models.BookRequest) (domain.Book, error) {
	return domain.NewBook(domain.NewBookData{
		Title:      bookRequest.Title,
		Year:       bookRequest.Year,
		Author:     bookRequest.Author,
		Price:      bookRequest.Price,
		Stock:      bookRequest.Stock,
		CategoryID: bookRequest.CategoryID,
	})
}

func ToDomainUser(username, password string) (domain.User, error) {
	return domain.NewUser(domain.NewUserData{
		Email:    username,
		Password: password,
	})
}

func ToDomainCart(userID int, cartRequest models.CartRequest) (domain.Cart, error) {
	return domain.NewCart(domain.NewCartData{
		UserID:  userID,
		BookIDs: cartRequest.BookIDs,
	})
}

func ToResponseCart(cart domain.Cart) models.CartResponse {
	return models.CartResponse{
		BookIDs: cart.BookIDs(),
	}
}

func GetUserFromContext(ctx context.Context) (domain.User, error) {
	contextUser := ctx.Value(ContextUserKey)
	if contextUser == nil {
		return domain.User{}, domain.ErrNoUserInContext
	}
	user, ok := contextUser.(domain.User)
	if !ok {
		return domain.User{}, domain.ErrNoUserInContext
	}
	return user, nil
}

func GetUserFromGRPCMetadata(ctx context.Context) (domain.User, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return domain.User{}, domain.ErrNoUserInContext
	}

	// Получаем значение "user-id" из метаданных
	userIDStr, ok := md["user-id"]
	if !ok || len(userIDStr) == 0 {
		return domain.User{}, domain.ErrInvalidUserID
	}

	// Парсим ID пользователя
	userID, err := strconv.Atoi(userIDStr[0])
	if err != nil {
		return domain.User{}, domain.ErrInvalidUserID
	}

	userEmail, ok := md["user-email"]
	if !ok || len(userEmail) == 0 {
		return domain.User{}, domain.ErrInvalidUserEmail
	}

	userAdminStr, ok := md["user-admin"]
	if !ok || len(userAdminStr) == 0 {
		return domain.User{}, domain.ErrMissingMetadata
	}
	userAdmin, err := strconv.ParseBool(userAdminStr[0])
	if err != nil {
		return domain.User{}, domain.ErrMissingMetadata
	}

	return domain.NewUserFromToken(domain.NewUserData{
		ID:    userID,
		Email: userEmail[0],
		Admin: userAdmin,
	})
}
