package httpserver

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"toptal/internal/app/common/server"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
)

func (s HttpServer) CheckAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(AuthorizationHeader)
		token, ok := extractBearerToken(token)
		if !ok {
			server.BadRequest("invalid-token", nil, w, r)
			return
		}

		user, err := s.authService.GetUserFromToken(token)
		if err != nil {
			server.Unauthorised("invalid-token", err, w, r)
			return
		}
		if !user.Admin() {
			server.Unauthorised("not-admin", nil, w, r)
			return
		}
		ctx := context.WithValue(r.Context(), ContextUserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s HttpServer) CheckAuthorizedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(AuthorizationHeader)
		token, ok := extractBearerToken(token)
		if !ok {
			server.BadRequest("invalid-token", nil, w, r)
			return
		}
		user, err := s.authService.GetUserFromToken(token)
		if err != nil {
			server.Unauthorised("invalid-token", err, w, r)
			return
		}
		ctx := context.WithValue(r.Context(), ContextUserKey, user)

		user, err = s.authService.GetUserFromToken(token)
		if err != nil {
			server.Unauthorised("invalid-token", err, w, r)
			return
		}
		// Pass metadata through for the gRPC Gateway
		r.Header.Set("user-id", strconv.Itoa(user.ID()))
		r.Header.Set("user-email", user.Email())
		r.Header.Set("user-admin", strconv.FormatBool(user.Admin()))

		ctx = context.WithValue(r.Context(), ContextUserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearerToken(token string) (string, bool) {
	if !strings.HasPrefix(token, BearerPrefix) {
		return "", false
	}
	token = strings.TrimPrefix(token, BearerPrefix)
	return token, token != ""

}
