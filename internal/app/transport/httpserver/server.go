package httpserver

import "toptal/internal/app/transport/interfaces"

type HttpServer struct {
	userService     interfaces.UserService
	authService     interfaces.AuthService
	bookService     interfaces.BookService
	cartService     interfaces.CartService
	categoryService interfaces.CategoryService
}

func NewHttpServer(userService interfaces.UserService,
	authService interfaces.AuthService,
	bookService interfaces.BookService,
	cartService interfaces.CartService,
	categoryService interfaces.CategoryService) *HttpServer {
	return &HttpServer{
		userService:     userService,
		authService:     authService,
		bookService:     bookService,
		cartService:     cartService,
		categoryService: categoryService,
	}
}
