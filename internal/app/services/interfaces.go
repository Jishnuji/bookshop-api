// go generate mockery
package services

import (
	"context"
	"toptal/internal/app/domain"
)

type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	CreateUser(ctx context.Context, data domain.User) (domain.User, error)
	GetUser(ctx context.Context, id int) (domain.User, error)
}

type BookRepository interface {
	GetBook(ctx context.Context, id int) (domain.Book, error)
	GetBooks(ctx context.Context, categoryIDs []int, limit, offset int) ([]domain.Book, error)
	CreateBook(ctx context.Context, book domain.Book) (domain.Book, error)
	UpdateBook(ctx context.Context, book domain.Book) (domain.Book, error)
	DeleteBook(ctx context.Context, id int) error
}

type CategoryRepository interface {
	CreateCategory(ctx context.Context, category domain.Category) (domain.Category, error)
	GetCategory(ctx context.Context, id int) (domain.Category, error)
	UpdateCategory(ctx context.Context, category domain.Category) (domain.Category, error)
	DeleteCategory(ctx context.Context, id int) error
	GetCategories(ctx context.Context) ([]domain.Category, error)
}

type CartRepository interface {
	GetCart(ctx context.Context, userID int) (domain.Cart, error)
	DeleteCart(ctx context.Context, userID int) error
	UpdateCartAndStocks(ctx context.Context, cart domain.Cart) error
	CheckStocks(ctx context.Context, cart domain.Cart) (bool, error)
}

type AuthRepository interface {
	Login(ctx context.Context, email, password string) (string, error)
	ValidateToken(ctx context.Context, token string) (domain.User, error)
}
