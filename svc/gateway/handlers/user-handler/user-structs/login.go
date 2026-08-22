package userstructs

type Login struct {
	Email    string `json:"email,case:ignore"`
	Password string `json:"password,case:ignore"`
}
