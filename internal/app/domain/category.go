package domain

import "fmt"

type Category struct {
	id   int
	name string
}

type NewCategoryData struct {
	ID   int
	Name string
}

// NewCategory constructs a Category from the provided data.
func NewCategory(data NewCategoryData) (Category, error) {
	if data.Name == "" {
		return Category{}, fmt.Errorf("%w: name", ErrRequired)
	}
	return Category{
		id:   data.ID,
		name: data.Name,
	}, nil
}

// ID returns the category identifier.
func (c Category) ID() int {
	return c.id
}

// Name returns the category name.
func (c Category) Name() string {
	return c.name
}
