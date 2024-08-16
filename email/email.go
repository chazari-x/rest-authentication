package email

import (
	"github.com/veyselaksin/gomailer/pkg/mailer"
	"rest-authentication/config"
	"rest-authentication/storage"
)

// Email represents the email service
type Email struct {
	cfg     config.Email
	sender  mailer.ISender
	storage *storage.Storage
}

// New creates a new Email instance
func New(cfg config.Email, store *storage.Storage) *Email {
	auth := mailer.Authentication{
		Username: cfg.Username,
		Password: cfg.Password,
		Host:     cfg.Host,
		Port:     cfg.Port,
	}

	sender := mailer.NewPlainAuth(&auth)

	return &Email{
		cfg:     cfg,
		sender:  sender,
		storage: store,
	}
}

// Send sends an email
func (e *Email) Send(GUID, subject, body string) (bool, error) {
	email, err := e.storage.SelectUserEmailByGUID(GUID)
	if err != nil {
		return false, err
	}

	m := mailer.NewMessage(subject, body)
	m.SetTo([]string{email})

	return true, e.sender.SendMail(m)
}
