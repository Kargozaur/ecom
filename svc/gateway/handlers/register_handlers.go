package handlers

import (
	userhandler "gateway/handlers/user-handler"
	"gateway/middleware"
	"net/http"
	"pkg/token"
	userv1 "proto/out/user/v1"
)

func RegisterUserHandler(mux *http.ServeMux, cl userv1.UserServiceClient, validator *token.TokenValidator) {
	mw := middleware.NewMiddleware(validator)
	handler := userhandler.NewUserHandler(cl)
	mux.HandleFunc("POST /user/register", handler.CreateUser)
	mux.HandleFunc("GET /user/get_profile", mw.ValidateToken(http.HandlerFunc(handler.GetProfile)))
}
