package userstructs

type Profile struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	RegisterData string `json:"registerDate"`
}
