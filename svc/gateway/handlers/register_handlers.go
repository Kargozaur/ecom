package handlers

import (
	userhandler "gateway/handlers/user-handler"
	"net/http"
	userv1 "proto/out/user/v1"
)

func RegisterUserHandler(mux *http.ServeMux, cl userv1.UserServiceClient) {
	handler := userhandler.NewUserHandler(cl)
	mux.HandleFunc("POST /user/register", handler.CreateUser)
}
