package httpserver

type HttpServer struct {
	userService     UserService
	authService     AuthService
	bookService     BookService
	cartService     CartService
	categoryService CategoryService
}

func NewHttpServer(userService UserService, authService AuthService, bookService BookService, cartService CartService, categoryService CategoryService) *HttpServer {
	return &HttpServer{
		userService:     userService,
		authService:     authService,
		bookService:     bookService,
		cartService:     cartService,
		categoryService: categoryService,
	}
}
