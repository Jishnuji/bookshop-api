package httpserver

import (
	"context"
	"toptal/internal/app/domain"
)

type UserService interface {
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	CreateUser(ctx context.Context, data domain.User) (domain.User, error)
	GetUser(ctx context.Context, id int) (domain.User, error)
}

type BookService interface {
	CreateBook(ctx context.Context, data domain.Book) (domain.Book, error)
	GetBook(ctx context.Context, in int) (domain.Book, error)
	GetBooks(ctx context.Context, categoryIDs []int, limit, offset int) ([]domain.Book, error)
	UpdateBook(ctx context.Context, book domain.Book) (domain.Book, error)
	DeleteBook(ctx context.Context, id int) error
}

type CategoryService interface {
	CreateCategory(ctx context.Context, category domain.Category) (domain.Category, error)
	GetCategory(ctx context.Context, id int) (domain.Category, error)
	UpdateCategory(ctx context.Context, category domain.Category) (domain.Category, error)
	DeleteCategory(ctx context.Context, id int) error
	GetCategories(ctx context.Context) ([]domain.Category, error)
}

type CartService interface {
	UpdateCartAndStocks(ctx context.Context, cart domain.Cart) (domain.Cart, error)
	Checkout(ctx context.Context, userID int) error
}

type AuthService interface {
	GetUserFromToken(token string) (domain.User, error)
	GenerateToken(user domain.User) (string, error)
}
