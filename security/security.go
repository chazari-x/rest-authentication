package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"rest-authentication/config"
	"rest-authentication/email"
	"rest-authentication/model"
	"rest-authentication/storage"
)

// Security represents the security service
type Security struct {
	cfg     config.Security
	storage *storage.Storage
	email   *email.Email
}

// New creates a new Security instance
func New(cfg config.Security, storage *storage.Storage, email *email.Email) *Security {
	return &Security{cfg: cfg, storage: storage, email: email}
}

// InsertUser inserts a new user into the database
func (s *Security) InsertUser(email, password string) (string, error) {
	user := model.User{
		GUID:     uuid.New().String(),
		Email:    email,
		Password: s.createHash(password),
	}
	return s.storage.InsertUser(user)
}

// SelectUserByGUIDAndPass selects a user by GUID and password
func (s *Security) SelectUserByGUIDAndPass(GUID, password string) (model.User, error) {
	return s.storage.SelectUserByGUIDAndPass(GUID, s.createHash(password))
}

// GenerateTokens generates access and refresh tokens
func (s *Security) GenerateTokens(GUID, IP string) (string, string, error) {
	UUID := uuid.New().String()

	if has, err := s.storage.HasUUIDToken(UUID); err != nil {
		return "", "", err
	} else if has {
		return s.GenerateTokens(GUID, IP)
	}

	access, err := s.generateAccessToken(GUID, IP, UUID)
	if err != nil {
		return "", "", err
	}

	refresh, err := s.generateRefreshToken(access)
	if err != nil {
		return "", "", err
	}

	return access, refresh, s.storage.InsertRefreshToken(UUID, s.createHash(refresh))
}

// RefreshTokens refreshes access and refresh tokens
func (s *Security) RefreshTokens(access, refresh, IP string) (string, string, error) {
	token, err := s.validateToken(access)
	if err != nil {
		return "", "", err
	}

	if !s.ValidateRefresh(access, refresh) {
		return "", "", nil
	}

	claims := token.Claims.(jwt.MapClaims)

	if has, err := s.storage.HasRefreshToken(claims["uuid"].(string), s.createHash(refresh)); err != nil || !has {
		return "", "", err
	}

	if claims["ip"].(string) != IP {
		err = s.email.Send(
			claims["guid"].(string),
			"Warning - Account Security",
			"WARNING: Your account has been accessed from a new IP address.",
		)
		if err != nil {
			log.Error(err)
			return "", "", err
		}
	}

	if err := s.storage.DeleteRefreshToken(claims["uuid"].(string)); err != nil {
		return "", "", err
	}

	return s.GenerateTokens(claims["guid"].(string), IP)
}

// generateAccessToken generates an access token
func (s *Security) generateAccessToken(GUID, IP, UUID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"uuid":      UUID,
		"guid":      GUID,
		"ip":        IP,
		"timestamp": time.Now(),
	})

	return token.SignedString([]byte(s.cfg.SecretKey))
}

// generateRefreshToken generates a refresh token
func (s *Security) generateRefreshToken(access string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"access": access,
	})

	signedToken, err := token.SignedString([]byte(s.cfg.SecretKey))
	if err != nil {
		return "", err
	}

	elements := strings.Split(signedToken, ".")
	refresh := base64.StdEncoding.EncodeToString([]byte(elements[len(elements)-1]))

	return refresh, nil
}

// ValidateAccess validates an access token
func (s *Security) ValidateAccess(access string) error {
	token, err := s.validateToken(access)
	if err != nil {
		return err
	}

	claims := token.Claims.(jwt.MapClaims)

	_, err = s.storage.HasUUIDToken(claims["uuid"].(string))
	if err != nil {
		return err
	}

	return nil
}

// validateToken validates a token
func (s *Security) validateToken(signedToken string) (*jwt.Token, error) {
	token, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

// ValidateRefresh validates a refresh token
func (s *Security) ValidateRefresh(access, encodeRefresh string) bool {
	decodeRefresh, err := base64.StdEncoding.DecodeString(encodeRefresh)
	if err != nil {
		return false
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"access": access,
	})

	signedToken, err := token.SignedString([]byte(s.cfg.SecretKey))
	if err != nil {
		return false
	}

	elements := strings.Split(signedToken, ".")
	newRefresh := elements[len(elements)-1]

	return newRefresh == string(decodeRefresh)
}

// createHash creates a hash from a text
func (s *Security) createHash(text string) string {
	sig := hmac.New(sha256.New, []byte(s.cfg.SecretKey))
	sig.Write([]byte(text))

	return hex.EncodeToString(sig.Sum(nil))
}
