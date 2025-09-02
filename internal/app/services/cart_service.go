// internal/app/services/cart_service.go
package services

import (
	"context"
	"fmt"
	"toptal/internal/app/domain"

	"github.com/davecgh/go-spew/spew"
)

type CartService struct {
	cartRepo CartRepository
}

// NewCartService creates a new cart service instance
func NewCartService(cartRepo CartRepository) *CartService {
	return &CartService{
		cartRepo: cartRepo,
	}
}

// UpdateCart updates a cart
func (s CartService) UpdateCartAndStocks(ctx context.Context, cart domain.Cart) (domain.Cart, error) {
	err := s.cartRepo.UpdateCartAndStocks(ctx, cart)
	if err != nil {
		return domain.Cart{}, fmt.Errorf("failed to update cart and stocks: %w", err)
	}

	updatedCart, err := s.cartRepo.GetCart(ctx, cart.UserID())
	spew.Dump(err)
	if err != nil {
		return domain.Cart{}, fmt.Errorf("failed to get updated cart: %w", err)
	}

	return updatedCart, nil
}

// Checkout does nothing but cleans up the cart as per the spec
func (s CartService) Checkout(ctx context.Context, userID int) error {
	return s.cartRepo.DeleteCart(ctx, userID)
}
