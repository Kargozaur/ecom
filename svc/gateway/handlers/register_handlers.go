package handlers

import (
	userhandler "gateway/handlers/user-handler"
	"gateway/middleware"
	"net/http"
	"pkg/credvalidator"
	"pkg/token"
	userv1 "proto/out/user/v1"
)

func RegisterUserHandler(mux *http.ServeMux, cl userv1.UserServiceClient, validator *token.TokenValidator, policies credvalidator.PasswordPolicy) {
	mw := middleware.NewMiddleware(validator)
	handler := userhandler.NewUserHandler(cl, policies)
	mux.HandleFunc("POST /user/register", handler.CreateUser)
	mux.HandleFunc("GET /user/get_profile", mw.SetToken(http.HandlerFunc(handler.GetProfile)))
	mux.HandleFunc("POST /user/login", handler.Login)

}
