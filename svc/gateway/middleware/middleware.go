package middleware

import (
	"context"
	"net/http"
	"pkg/token"
	"strings"
)

type ContextKey string

const ClaimsKey ContextKey = "jwtClaims"

type middleware struct {
	validator token.ITokenValidator
}

func NewMiddleware(validator token.ITokenValidator) *middleware {
	return &middleware{validator: validator}
}

func (m *middleware) SetToken(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jwtToken string
		if cookie, err := r.Cookie("access_token"); err == nil {
			jwtToken = cookie.Value
		} else {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "failed to get a token", http.StatusUnauthorized)
				return
			}
			cut, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok {
				http.Error(w, "wrong token format", http.StatusBadRequest)
				return
			}
			jwtToken = cut
		}
		ctx := context.WithValue(r.Context(), ClaimsKey, jwtToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *middleware) SetUserID(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cl := r.Context().Value(ClaimsKey)
		jwtToken, ok := cl.(string)
		if !ok {
			http.Error(w, "invalid token", http.StatusBadRequest)
			return
		}
		claims, err := m.validator.GetClaims(jwtToken, token.Access)
		if err != nil {
			http.Error(w, "invalid token", http.StatusBadRequest)
			return
		}
		r.Header.Set("userID", claims.UserID)
		next.ServeHTTP(w, r)
	})
}
