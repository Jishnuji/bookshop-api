package httpserver

import (
	"encoding/json"
	"net/http"
	"toptal/internal/app/common/server"
	"toptal/internal/app/transport/models"
)

func (s HttpServer) SignUp(w http.ResponseWriter, r *http.Request) {
	var authRequest models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&authRequest); err != nil {
		server.BadRequest("invalid-json", err, w, r)
	}

	authRequest.Normalize()

	hashedPassword, err := hashPassword(authRequest.Password)
	if err != nil {
		server.RespondWithError(err, w, r)
	}

	user, err := toDomainUser(authRequest.Email, hashedPassword)
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}
	_, err = s.userService.CreateUser(r.Context(), user)
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	server.RespondOK(map[string]bool{"ok": true}, w, r)
}

func (s HttpServer) SignIn(w http.ResponseWriter, r *http.Request) {
	var authRequest models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&authRequest); err != nil {
		server.BadRequest("invalid-json", err, w, r)
	}

	user, err := s.userService.GetUserByEmail(r.Context(), authRequest.Email)
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	if !checkPasswordHash(authRequest.Password, user.Password()) {
		server.BadRequest("invalid-password", nil, w, r)
		return
	}

	token, err := s.authService.GenerateToken(user)
	if err != nil {
		server.RespondWithError(err, w, r)
		return
	}

	server.RespondOK(map[string]string{"token": token}, w, r)
}
