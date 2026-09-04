package userhandler

import (
	"errors"
	userstructs "gateway/handlers/user-handler/user-structs"
	"gateway/types"
	"net/http"
	"pkg/credvalidator"
	"pkg/json"
	userv1 "proto/out/user/v1"
	"time"
)

type UserHandler struct {
	userClient userv1.UserServiceClient
	Policies   credvalidator.PasswordPolicy
}

func NewUserHandler(cl userv1.UserServiceClient, p credvalidator.PasswordPolicy) *UserHandler {
	return &UserHandler{userClient: cl, Policies: p}
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
		http.Error(w, "failed to register", http.StatusRequestTimeout)
		return
	}
	json.Write(w, http.StatusCreated, map[string]string{"response": resp.GetResponse()})
}

func (u *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	token := r.Context().Value(types.TokenKey).(string)
	data := &userv1.GetProfileRequest{Jwt: token}
	req, err := u.userClient.GetProfile(r.Context(), data)
	if err != nil {
		http.Error(w, "failed to fetch profile", http.StatusBadRequest)
		return
	}
	responseBody := &userstructs.Profile{
		Email:        req.GetEmail(),
		Name:         req.GetName(),
		RegisterData: req.GetDate(),
	}
	json.Write(w, http.StatusOK, responseBody)
}

func (u *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body userstructs.Login
	if err := json.Read(r, &body); err != nil {
		http.Error(w, "Failed to parse request body "+err.Error(), http.StatusBadRequest)
		return
	}
	data := &userv1.LoginUserRequest{Email: body.Email, Password: body.Password}
	resp, err := u.userClient.LoginUser(r.Context(), data)
	if err != nil {
		http.Error(w, "failed to login", http.StatusBadRequest)
		return
	}
	access := &http.Cookie{
		Name:     "access_token",
		Value:    resp.GetAccess(),
		Path:     "/",
		Expires:  time.Now().UTC().Add(time.Minute * 15),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	refresh := &http.Cookie{
		Name:     "refresh_token",
		Value:    resp.GetRefresh(),
		Path:     "/",
		Expires:  time.Now().UTC().Add(time.Hour * 24 * 7),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, access)
	http.SetCookie(w, refresh)
	json.Write(w, http.StatusOK,
		map[string]string{"token": resp.GetAccess(), "type": resp.GetType()})
}

func (u *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "failed to get a cookie", http.StatusBadRequest)
		return
	}
	data := &userv1.LogoutRequest{RefreshToken: refreshCookie.Value}
	res, err := u.userClient.Logout(r.Context(), data)
	if err != nil {
		http.Error(w, "failed to log out", http.StatusBadRequest)
		return
	}
	access := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Now().Add(time.Hour * -1),
	}
	refresh := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Now().Add(-time.Hour),
	}
	http.SetCookie(w, access)
	http.SetCookie(w, refresh)
	json.Write(w, http.StatusAccepted, map[string]string{
		"response": res.GetResponse(),
	})
}
