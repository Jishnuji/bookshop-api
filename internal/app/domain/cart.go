package domain

import (
	"fmt"
	"slices"
)

type Cart struct {
	userID  int
	bookIDs []int
}

type NewCartData struct {
	UserID  int
	BookIDs []int
}

// NewCart constructs a Cart from the provided data.
func NewCart(data NewCartData) (Cart, error) {
	if data.UserID == 0 {
		return Cart{}, fmt.Errorf("%w: user_id", ErrInvalidUserID)
	}
	if len(data.BookIDs) == 0 {
		return Cart{}, fmt.Errorf("%w: book_ids", ErrNil)
	}

	uniqueBookIDs, err := removeDuplicates(data.BookIDs)
	if err != nil {
		return Cart{}, err
	}
	return Cart{
		userID:  data.UserID,
		bookIDs: uniqueBookIDs,
	}, nil

}

func removeDuplicates(bookIDs []int) ([]int, error) {
	seen := make(map[int]struct{}, len(bookIDs))
	unique := make([]int, 0, len(bookIDs))
	for _, id := range bookIDs {
		_, exists := seen[id]
		if id <= 0 {
			return nil, fmt.Errorf("%w: bookID", ErrNegative)
		}
		if !exists {
			seen[id] = struct{}{}
			unique = append(unique, id)
		}
	}

	return unique, nil
}

// UserID returns the cart owner's user identifier.
func (c *Cart) UserID() int {
	return c.userID
}

// BookIDs returns the map of book IDs currently in the cart.
func (c *Cart) BookIDs() []int {
	return c.bookIDs
}

// AddBook adds a book to the cart by its ID.
func (c *Cart) AddBook(bookID int) {
	if !c.HasBook(bookID) {
		c.bookIDs = append(c.bookIDs, bookID)
	}
}

// RemoveBook removes a book from the cart by its ID.
func (c *Cart) RemoveBook(bookID int) {
	for i, id := range c.bookIDs {
		if id == bookID {
			c.bookIDs = slices.Delete(c.bookIDs, i, i+1)
			break
		}
	}
}

// Clear removes all books from the cart.
func (c *Cart) Clear() {
	c.bookIDs = c.bookIDs[:0]
}

// HasBook checks if a book with the given ID exists in the cart.
func (c *Cart) HasBook(bookID int) bool {
	return slices.Contains(c.bookIDs, bookID)
}

// HasBooks checks if the cart contains any books.
func (c *Cart) HasBooks() bool {
	return len(c.bookIDs) > 0
}

// Equal compares two carts for equality.
func (c *Cart) Equal(other Cart) bool {
	if c.UserID() != other.UserID() {
		return false
	}

	if len(c.BookIDs()) != len(other.BookIDs()) {
		return false
	}

	for _, bookID := range other.BookIDs() {
		if !c.HasBook(bookID) {
			return false
		}
	}

	return true
}

// Diff returns a new cart containing only the books that exist in this cart
func (c *Cart) Diff(old Cart) Cart {
	diff, _ := NewCart(NewCartData{
		UserID: c.UserID(),
	})

	for _, bookID := range c.BookIDs() {
		if !old.HasBook(bookID) {
			diff.AddBook(bookID)
		}
	}

	return diff
}

// Merge combines books from this cart and the old cart into a new cart.
func (c *Cart) Merge(old Cart) Cart {
	totalSize := len(c.BookIDs()) + len(old.BookIDs())
	allBooks := make([]int, 0, totalSize)
	allBooks = append(allBooks, c.bookIDs...)
	allBooks = append(allBooks, old.bookIDs...)

	merged, _ := NewCart(NewCartData{
		UserID:  c.UserID(),
		BookIDs: allBooks,
	})

	return merged
}
