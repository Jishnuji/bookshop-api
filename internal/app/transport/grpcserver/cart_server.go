package grpcserver

import (
	"context"
	"errors"
	"toptal/internal/app/common/auth"
	"toptal/internal/app/domain"
	"toptal/internal/app/transport/interfaces"
	cartv1 "toptal/proto/v1/cart"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CartServer struct {
	cartv1.UnimplementedCartServiceServer
	cartService interfaces.CartService
	userService interfaces.UserService
}

func NewCartServer(cartService interfaces.CartService, userService interfaces.UserService) *CartServer {
	return &CartServer{
		cartService: cartService,
		userService: userService,
	}
}

func (s *CartServer) UpdateCart(ctx context.Context, req *cartv1.UpdateCartRequest) (*cartv1.UpdateCartResponse, error) {
	user, err := auth.GetUserFromGRPCMetadata(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	// Verify the user exists
	_, err = s.userService.GetUser(ctx, user.ID())
	if err != nil {
		return nil, toSlugError(err)
	}

	// Convert the gRPC request to the domain model
	domainCart, err := toDomainCartFromGRPC(user.ID(), req.Cart)
	if err != nil {
		if errors.Is(err, domain.ErrNegative) {
			return nil, status.Error(codes.InvalidArgument, "invalid book_id")
		}
		if errors.Is(err, domain.ErrNil) {
			return nil, status.Error(codes.InvalidArgument, "missing book_ids")
		}
		if errors.Is(err, domain.ErrInvalidUserID) {
			return nil, status.Error(codes.InvalidArgument, "invalid user_id")
		}
		return nil, toSlugError(err)
	}

	// Update the cart via the service
	updatedCart, err := s.cartService.UpdateCartAndStocks(ctx, domainCart)
	if err != nil {
		return nil, toSlugError(err)
	}

	return &cartv1.UpdateCartResponse{
		UserId: int64(updatedCart.UserID()),
		Cart:   toGRPCCartData(updatedCart),
	}, nil
}

func (s *CartServer) Checkout(ctx context.Context, req *cartv1.CheckoutRequest) (*cartv1.CheckoutResponse, error) {
	// Get the user from the context
	user, err := auth.GetUserFromGRPCMetadata(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	// Place the order via the service
	err = s.cartService.Checkout(ctx, user.ID())
	if err != nil {
		return nil, toSlugError(err)
	}

	return &cartv1.CheckoutResponse{
		Success: true,
	}, nil
}
