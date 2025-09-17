package grpcserver

import (
	"fmt"
	"net"
	"toptal/internal/app/transport/interfaces"
	authv1 "toptal/proto/v1/auth"
	"toptal/proto/v1/book"
	cartv1 "toptal/proto/v1/cart"
	categoryv1 "toptal/proto/v1/category"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcServer struct {
	userService     interfaces.UserService
	authService     interfaces.AuthService
	bookService     interfaces.BookService
	cartService     interfaces.CartService
	categoryService interfaces.CategoryService
	server          *grpc.Server
}

func NewGrpcServer(userService interfaces.UserService,
	authService interfaces.AuthService,
	bookService interfaces.BookService,
	cartService interfaces.CartService,
	categoryService interfaces.CategoryService,

) *GrpcServer {
	return &GrpcServer{
		userService:     userService,
		authService:     authService,
		bookService:     bookService,
		cartService:     cartService,
		categoryService: categoryService,
	}
}

func (s *GrpcServer) Start(grpcAddr string) error {
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.server = grpc.NewServer()

	// Register services
	s.registerServices(s.server)

	// Enable reflection for grpcurl/gRPC clients
	reflection.Register(s.server)

	return s.server.Serve(listener)
}

func (s *GrpcServer) registerServices(server *grpc.Server) {
	// Register AuthService
	authServer := NewAuthServer(s.userService, s.authService)
	bookServer := NewBookServer(s.bookService)
	categoryServer := NewCategoryServer(s.categoryService)
	cartServer := NewCartServer(s.cartService, s.userService)
	authv1.RegisterAuthServiceServer(server, authServer)
	bookv1.RegisterBookServiceServer(server, bookServer)
	categoryv1.RegisterCategoryServiceServer(server, categoryServer)
	cartv1.RegisterCartServiceServer(server, cartServer)
}

func (s *GrpcServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}
