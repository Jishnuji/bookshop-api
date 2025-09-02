package models

type CartRequest struct {
	BookIDs []int `json:"book_ids"`
}

type CartResponse struct {
	BookIDs []int `json:"book_ids"`
}
