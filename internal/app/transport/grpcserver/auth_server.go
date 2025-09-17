package grpcserver

import (
	"context"
	"strings"
	"toptal/internal/app/common/auth"
	"toptal/internal/app/transport/interfaces"
	authv1 "toptal/proto/v1/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	userService interfaces.UserService
	authService interfaces.AuthService
}

func NewAuthServer(userService interfaces.UserService, authService interfaces.AuthService) *AuthServer {
	return &AuthServer{
		userService: userService,
		authService: authService,
	}
}

func (s *AuthServer) SignUp(ctx context.Context, req *authv1.SignUpRequest) (*authv1.SignUpResponse, error) {
	// Та же логика, что в SignUp handler
	// Валидация входных данных
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	// Нормализация email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Хеширование пароля
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	// Создание доменного пользователя
	user, err := auth.ToDomainUser(req.Email, hashedPassword)
	if err != nil {
		return nil, toSlugError(err)
	}

	// Создание пользователя через сервис
	_, err = s.userService.CreateUser(ctx, user)
	if err != nil {
		return nil, toSlugError(err)
	}

	return &authv1.SignUpResponse{Success: true}, nil
}

func (s *AuthServer) SignIn(ctx context.Context, req *authv1.SignInRequest) (*authv1.SignInResponse, error) {
	// Получение пользователя по email
	user, err := s.userService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, toSlugError(err)
	}

	// Проверка пароля
	if !auth.CheckPasswordHash(req.Password, user.Password()) {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	// Генерация токена
	token, err := s.authService.GenerateToken(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &authv1.SignInResponse{Token: token}, nil
}
