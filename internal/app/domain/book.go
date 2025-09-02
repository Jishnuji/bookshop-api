package domain

import "fmt"

// Book is a domain book.
type Book struct {
	id         int
	title      string
	year       int
	author     string
	price      int
	stock      int
	categoryID int
}

type NewBookData struct {
	ID         int
	Title      string
	Year       int
	Author     string
	Price      int
	Stock      int
	CategoryID int
}

func NewBook(data NewBookData) (Book, error) {
	if err := validateBookData(data); err != nil {
		return Book{}, fmt.Errorf("faild book data validation: %w", err)
	}
	return Book{
		id:         data.ID,
		title:      data.Title,
		year:       data.Year,
		author:     data.Author,
		price:      data.Price,
		stock:      data.Stock,
		categoryID: data.CategoryID,
	}, nil
}

func validateBookData(data NewBookData) error {
	if data.Title == "" {
		return fmt.Errorf("%w: title", ErrRequired)
	}
	if data.Year <= 0 {
		return fmt.Errorf("%w: year", ErrNegative)
	}
	if data.Author == "" {
		return fmt.Errorf("%w: author", ErrRequired)
	}
	if data.Price <= 0 {
		return fmt.Errorf("%w: price", ErrNegative)
	}
	if data.Stock < 0 {
		return fmt.Errorf("%w: stock", ErrNegative)
	}
	if data.CategoryID == 0 {
		return fmt.Errorf("%w: category_id", ErrRequired)
	}
	return nil
}

func (b Book) ID() int {
	return b.id
}

func (b Book) Title() string {
	return b.title
}

func (b Book) Year() int {
	return b.year
}

func (b Book) Author() string {
	return b.author
}

func (b Book) Price() int {
	return b.price
}

func (b Book) Stock() int {
	return b.stock
}

func (b Book) CategoryID() int {
	return b.categoryID
}
