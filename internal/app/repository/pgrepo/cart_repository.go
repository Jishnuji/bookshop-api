package pgrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"toptal/internal/app/common/slugerrors"
	"toptal/internal/app/domain"
	"toptal/internal/app/repository/models"
	"toptal/internal/pkg/pg"

	"github.com/uptrace/bun"
)

type CartRepository struct {
	db *pg.DB
}

// NewCartRepository creates a new cart repository instance
func NewCartRepository(db *pg.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r CartRepository) GetCart(ctx context.Context, userID int) (domain.Cart, error) {
	var cart models.Cart
	err := r.db.NewSelect().Model(&cart).Where("user_id = ?", userID).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Cart{}, domain.ErrNotFound
		}
		return domain.Cart{}, fmt.Errorf("failed to get cart: %w", err)
	}

	domainCart, err := cartToDomain(cart)
	if err != nil {
		return domain.Cart{}, fmt.Errorf("failed to create domain cart: %w", err)
	}

	return domainCart, nil
}

func (r CartRepository) UpdateCartAndStocks(ctx context.Context, cart domain.Cart) error {
	err := pg.HandleBunTransaction(ctx, func(tx bun.Tx) error {
		oldCart, err := r.GetCart(ctx, cart.UserID())
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("failed to get cart: %w", err)
		}

		if cart.Equal(oldCart) {
			return nil
		}

		var cartAdd, cartRemove, cartUnion domain.Cart
		if oldCart.HasBooks() {
			cartAdd = cart.Diff(oldCart)
			cartRemove = oldCart.Diff(cart)
			cartUnion = oldCart.Merge(cart)
		} else {
			cartAdd = cart
			cartUnion = cart
		}

		ok, err := r.CheckStocks(ctx, cartAdd)
		if err != nil {
			return fmt.Errorf("failed to check stocks: %w", err)
		}
		if !ok {
			return slugerrors.NewBadRequestError("some books are out of stock", "out-of-stock")
		}

		var dbStocks []models.Book //Used for locking stocks
		if cartUnion.HasBooks() {
			err := tx.NewRaw("SELECT id, stock FROM ? where id in (?) FOR UPDATE", bun.Ident("books"), bun.In(cartUnion.BookIDs())).Scan(ctx, &dbStocks)
			if err != nil {
				return fmt.Errorf("failed to lock stocks: %w", err)
			}
		}

		if cartAdd.HasBooks() {
			_, err := tx.NewUpdate().Model((*models.Book)(nil)).Set("stock = stock - 1").Where("id in (?)", bun.In(cartAdd.BookIDs())).Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to reduce stock: %w", err)
			}
		}
		if cartRemove.HasBooks() {
			_, err := tx.NewUpdate().Model((*models.Book)(nil)).Set("stock = stock + 1").Where("id in (?)", bun.In(cartRemove.BookIDs())).Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to add stock: %w", err)
			}
		}

		dbCart := domainToCart(cart)
		dbCart.UpdatedAt = time.Now()

		err = tx.NewInsert().Model(&dbCart).
			On("CONFLICT (user_id) DO UPDATE").
			Set("book_ids = EXCLUDED.book_ids").
			Set("updated_at = EXCLUDED.updated_at").
			Scan(ctx)
		if err != nil {
			return fmt.Errorf("failed to update cart: %w", err)
		}

		return nil
	}, r.db.DB)
	if err != nil {
		return fmt.Errorf("failed to update cart and stock: %w", err)
	}

	return nil
}

func (r CartRepository) CheckStocks(ctx context.Context, cart domain.Cart) (bool, error) {
	var books []models.Book
	err := r.db.NewSelect().Model(&books).Where("id in (?)", bun.In(cart.BookIDs())).Scan(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get stocks: %w", err)
	}

	stockMap := make(map[int]int)
	for _, book := range books {
		stockMap[book.ID] = book.Stock
	}

	for _, bookID := range cart.BookIDs() {
		if stockMap[bookID] == 0 {
			return false, nil
		}
	}

	return true, nil
}

// DeleteCart deletes a cart
func (r CartRepository) DeleteCart(ctx context.Context, userID int) error {
	_, err := r.db.NewDelete().Model((*models.Cart)(nil)).Where("user_id = ?", userID).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete cart: %w", err)
	}

	return nil
}

func (r CartRepository) CleanExpiredCarts(ctx context.Context, ttl time.Duration) error {
	err := pg.HandleBunTransaction(ctx, func(tx bun.Tx) error {
		var expiredCarts []models.Cart
		err := tx.NewSelect().Model(&expiredCarts).Where("updated_at < ?", time.Now().Add(-ttl)).Scan(ctx)
		if err != nil {
			return fmt.Errorf("failed to get expired carts: %w", err)
		}

		for _, cart := range expiredCarts {
			for _, bookID := range cart.BookIDs {
				_, err := tx.NewUpdate().Model((*models.Book)(nil)).Set("stock = stock + 1").Where("id = ?", bookID).Exec(ctx)
				if err != nil {
					return fmt.Errorf("failed to return stock: %w", err)
				}
			}
			_, err := tx.NewDelete().Model(&cart).Where("user_id = ?", cart.UserID).Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to delete cart: %w", err)
			}
		}

		return nil
	}, r.db.DB)
	if err != nil {
		return fmt.Errorf("failed to clean expired carts: %w", err)
	}

	return nil
}
