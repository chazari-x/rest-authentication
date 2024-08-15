package config

type DB struct {
	Addr string `validate:"required"`
}

type Server struct {
	Address string `validate:"required"`
}

type Security struct {
	SecretKey string `validate:"required"`
}

type Email struct {
	Host     string `validate:"required"`
	Port     string `validate:"required,numeric"`
	Username string `validate:"required"`
	Password string `validate:"required"`
}
