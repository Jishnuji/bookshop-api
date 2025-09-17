package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	auth "toptal/internal/app/common/auth"
	"toptal/internal/app/common/server"
	"toptal/internal/app/domain"
	"toptal/internal/app/transport/models"
)

func (s HttpServer) UpdateCart(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(r.Context())
	if err != nil {
		server.BadRequest("invalid-user", err, w, r)
		return
	}

	var cartRequest models.CartRequest

	if err := json.NewDecoder(r.Body).Decode(&cartRequest); err != nil {
		server.BadRequest("invalid-json", err, w, r)
		return
	}

	_, err = s.userService.GetUser(r.Context(), user.ID())
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	cart, err := auth.ToDomainCart(user.ID(), cartRequest)
	if err != nil {
		if errors.Is(err, domain.ErrNegative) {
			server.BadRequest("invalid-book-id", err, w, r)
			return
		}
		if errors.Is(err, domain.ErrNil) {
			server.BadRequest("missing-book-ids", err, w, r)
			return
		}
		if errors.Is(err, domain.ErrInvalidUserID) {
			server.BadRequest("invalid-user-id", err, w, r)
			return
		}
		server.RespondWithError(err, w, r)
		return
	}

	updatedCart, err := s.cartService.UpdateCartAndStocks(r.Context(), cart)
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	response := auth.ToResponseCart(updatedCart)

	server.RespondOK(response, w, r)
}

func (s HttpServer) Checkout(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(r.Context())
	if err != nil {
		server.BadRequest("invalid-user", err, w, r)
		return
	}

	err = s.cartService.Checkout(r.Context(), user.ID())
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	server.RespondOK(map[string]bool{"ok": true}, w, r)
}
