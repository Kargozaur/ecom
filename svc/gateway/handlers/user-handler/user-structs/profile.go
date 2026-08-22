package userstructs

type Profile struct {
	Name         string `json:"name,case:ignore"`
	Email        string `json:"email,case:ignore"`
	RegisterData string `json:"registerDate,case:ignore"`
}
