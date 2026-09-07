package handlers

import (
	userhandler "gateway/handlers/user-handler"
	"gateway/middleware"
	"net/http"
	"pkg/credvalidator"
	userv1 "proto/out/user/v1"
)

func RegisterUserHandler(mux *http.ServeMux, cl userv1.UserServiceClient, policies credvalidator.PasswordPolicy, mw *middleware.Middleware) {
	handler := userhandler.NewUserHandler(cl, policies)
	mux.HandleFunc("POST /user/register", handler.CreateUser)
	mux.HandleFunc("GET /user/get_profile", mw.SetToken(http.HandlerFunc(handler.GetProfile)))
	mux.HandleFunc("POST /user/login", handler.Login)
	mux.HandleFunc("POST /user/logout", mw.SetToken(http.HandlerFunc(handler.Logout)))
}
