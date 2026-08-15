package userhandler

import (
	"errors"
	userstructs "gateway/handlers/user-handler/user-structs"
	"gateway/middleware"
	"net/http"
	"pkg/credvalidator"
	"pkg/json"
	userv1 "proto/out/user/v1"
)

type UserHandler struct {
	userClient userv1.UserServiceClient
	Policies   *credvalidator.PasswordPolicy
}

func NewUserHandler(cl userv1.UserServiceClient, p *credvalidator.PasswordPolicy) *UserHandler {
	return &UserHandler{userClient: cl,
		Policies: p}
}

func (u *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var body userstructs.Register
	if err := json.Read(r, &body); err != nil {
		http.Error(w, "Failed to parse request body "+err.Error(), http.StatusBadRequest)
		return
	}
	errs := body.ValidateData(u.Policies)
	if len(errs) != 0 {
		err := errors.Join(errs...)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data := &userv1.RegisterUserRequest{Email: body.Email, Password: body.Password, Name: body.Name}
	resp, err := u.userClient.RegisterUser(r.Context(), data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusRequestTimeout)
		return
	}
	json.Write(w, http.StatusCreated, map[string]string{"response": resp.GetResponse()})
}

func (u *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value(middleware.ClaimsKey).(string)
	data := &userv1.GetProfileRequest{Jwt: token}
	req, err := u.userClient.GetProfile(r.Context(), data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	responseBody := &userstructs.Profile{
		Email:        req.GetEmail(),
		Name:         req.GetName(),
		RegisterData: req.GetDate(),
	}
	json.Write(w, http.StatusOK, responseBody)
}
