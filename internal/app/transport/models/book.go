package models

type BookRequest struct {
	Title      string `json:"title"`
	Year       int    `json:"year"`
	Author     string `json:"author"`
	Price      int    `json:"price"`
	Stock      int    `json:"stock"`
	CategoryID int    `json:"category_id"`
}

type BookResponse struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Year       int    `json:"year"`
	Author     string `json:"author"`
	Price      int    `json:"price"`
	Stock      int    `json:"stock"`
	CategoryID int    `json:"category_id"`
}
