package model

type RefreshToken struct {
	UUID    string `pg:",pk"`
	Refresh string `pg:",unique"`
}

type User struct {
	GUID     string `json:"guid" pg:",pk"`
	Email    string `json:"email" pg:",unique"`
	Password string `json:"password"`
}
