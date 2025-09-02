package models

import "strings"

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *AuthRequest) Normalize() {
	a.Email = strings.ToLower(strings.TrimSpace(a.Email))
}
