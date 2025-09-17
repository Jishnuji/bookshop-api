package httpserver

import (
	"context"
	"toptal/internal/app/domain"
	"toptal/internal/app/transport/models"
)

// Deprecated: use auth.ToResponseBook
func toResponseBook(book domain.Book) models.BookResponse {
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

// Deprecated: use auth.ToResponseCategory
func toResponseCategory(category domain.Category) models.CategoryResponse {
	return models.CategoryResponse{
		ID:   category.ID(),
		Name: category.Name(),
	}
}

// Deprecated: use auth.ToDomainBook
func toDomainBook(bookRequest models.BookRequest) (domain.Book, error) {
	return domain.NewBook(domain.NewBookData{
		Title:      bookRequest.Title,
		Year:       bookRequest.Year,
		Author:     bookRequest.Author,
		Price:      bookRequest.Price,
		Stock:      bookRequest.Stock,
		CategoryID: bookRequest.CategoryID,
	})
}

// Deprecated: use auth.ToDomainUser
func toDomainUser(username, password string) (domain.User, error) {
	return domain.NewUser(domain.NewUserData{
		Email:    username,
		Password: password,
	})
}

// Deprecated: use auth.ToDomainCart
func toDomainCart(userID int, cartRequest models.CartRequest) (domain.Cart, error) {
	return domain.NewCart(domain.NewCartData{
		UserID:  userID,
		BookIDs: cartRequest.BookIDs,
	})
}

// Deprecated: use auth.ToResponseCart
func toResponseCart(cart domain.Cart) models.CartResponse {
	return models.CartResponse{
		BookIDs: cart.BookIDs(),
	}
}

// Deprecated: use auth.GetUserFromContext
func getUserFromContext(ctx context.Context) (domain.User, error) {
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
