package services

import (
	"errors"
	"fmt"
	"time"
	"toptal/internal/app/domain"

	"github.com/golang-jwt/jwt/v5"
)

// TODO: move to secrets
var jwtSecretKey = []byte("f8e7d6c5b4a39281706f5e4d3c2b1a09887766554433221100ffeeddccbbaa99")

type AuthService struct {
	userRepo UserRepository
}

type UserClaims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Admin  bool   `json:"admin"`
	jwt.RegisteredClaims
}

// NewAuthService creates a new auth service instance
func NewAuthService(userRepo UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// GenerateToken generates a JWT token for a user
func (s *AuthService) GenerateToken(user domain.User) (string, error) {
	payload := UserClaims{
		UserID: user.ID(),
		Email:  user.Email(),
		Admin:  user.Admin(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	t, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return t, nil
}

// GetUserFromToken returns a user associated with a token
func (a *AuthService) GetUserFromToken(token string) (domain.User, error) {
	var userClaims UserClaims
	t, err := jwt.ParseWithClaims(token, &userClaims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecretKey, nil
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to parse a token: %w", err)
	}
	if !t.Valid {
		return domain.User{}, errors.New("invalid token")
	}
	user, err := userClaimsToDomainUser(userClaims)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to convert user claims to domain user: %w", err)
	}
	return user, nil
}

func userClaimsToDomainUser(claims UserClaims) (domain.User, error) {
	return domain.NewUserFromToken(domain.NewUserData{
		ID:    claims.UserID,
		Email: claims.Email,
		Admin: claims.Admin,
	})
}
