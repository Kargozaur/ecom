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

func (m *middleware) ValidateToken(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jwtToken string
		cookie, err := r.Cookie("access_token")
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if cookie != nil {
			jwtToken = cookie.String()
		} else {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "failed to get a token", http.StatusUnauthorized)
				return
			}
		}
		tokenString, ok := strings.CutPrefix(jwtToken, "Bearer ")
		if !ok {
			http.Error(w, "wrong token format", http.StatusBadRequest)
			return
		}
		claims, err := m.validator.ValidateToken(tokenString, token.Access)
		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *middleware) SetUserID(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cl := r.Context().Value(ClaimsKey)
		claims, ok := cl.(*token.Claims)
		if !ok {
			http.Error(w, token.ErrInvalidClaims.Error(), http.StatusBadRequest)
			return
		}
		r.Header.Set("userID", claims.UserID)
		next.ServeHTTP(w, r)
	})
}

func (m *middleware) SetEmail(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cl := r.Context().Value(ClaimsKey)
		claims, ok := cl.(*token.Claims)
		if !ok {
			http.Error(w, token.ErrInvalidClaims.Error(), http.StatusBadRequest)
			return
		}
		r.Header.Set("email", claims.Email)
		next.ServeHTTP(w, r)
	})
}
