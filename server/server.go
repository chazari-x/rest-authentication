package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"rest-authentication/config"
	_ "rest-authentication/docs"
	"rest-authentication/security"
)

// Server represents the server
type Server struct {
	cfg       config.Server
	security  *security.Security
	validator *validator.Validate
}

// New creates a new server
func New(cfg config.Server, sec *security.Security) *Server {
	return &Server{
		cfg:       cfg,
		security:  sec,
		validator: validator.New(),
	}
}

// Start starts the server
func (s *Server) Start() error {
	r := chi.NewRouter()

	r.Get("/api/swagger/*", httpSwagger.WrapHandler)

	r.Group(func(r chi.Router) {
		r.Use(s.middleware)

		r.Post("/api/auth", s.auth)

		r.Post("/api/register", s.register)

		r.Group(func(r chi.Router) {
			r.Use(s.validateAccessMiddleware)

			r.Get("/api/refresh", s.refresh)
		})
	})

	log.Info("server started on " + s.cfg.Address)
	return http.ListenAndServe(s.cfg.Address, r)
}

type responseTokens struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

type responseNewUser struct {
	GUID string `json:"guid"`
}

type responseError struct {
	Error string `json:"error"`
}

type requestNewUser struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type requestAuth struct {
	GUID     string `json:"guid" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// write writes a response
func write(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// @Summary Регистрация
// @Description Зарегистрировать нового пользователя
// @Produce  json
// @Param  message  body  requestNewUser  true  "Message content"
// @Success 201 {object} responseNewUser
// @Failure 400 {object} responseError
// @Failure 500 {object} responseError
// @Router /api/register [post]
func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		write(w, http.StatusBadRequest, responseError{Error: http.StatusText(http.StatusBadRequest)})
		return
	}

	var user requestNewUser
	if err = json.Unmarshal(body, &user); err != nil {
		write(w, http.StatusBadRequest, responseError{Error: http.StatusText(http.StatusBadRequest)})
		return
	}

	if err = s.validator.Struct(user); err != nil {
		write(w, http.StatusBadRequest, responseError{Error: err.Error()})
		return
	}

	guid, err := s.security.InsertUser(user.Email, user.Password)
	if err != nil {
		write(w, http.StatusInternalServerError, responseError{Error: http.StatusText(http.StatusInternalServerError)})
		return
	}

	if guid == "" {
		write(w, http.StatusBadRequest, responseError{Error: http.StatusText(http.StatusBadRequest)})
		return
	}

	write(w, http.StatusCreated, responseNewUser{GUID: guid})
}

// @Summary Аутентификация
// @Description Аутентификация пользователя
// @Produce  json
// @Param  message  body  requestAuth  true  "Message content"
// @Success 201 {object} responseTokens
// @Failure 400 {object} responseError
// @Failure 500 {object} responseError
// @Router /api/auth [post]
func (s *Server) auth(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		write(w, http.StatusBadRequest, responseError{Error: http.StatusText(http.StatusBadRequest)})
		return
	}

	var reqAuth requestAuth
	if err = json.Unmarshal(body, &reqAuth); err != nil {
		write(w, http.StatusBadRequest, responseError{Error: http.StatusText(http.StatusBadRequest)})
		return
	}

	if err = s.validator.Struct(reqAuth); err != nil {
		write(w, http.StatusBadRequest, responseError{Error: err.Error()})
		return
	}

	user, err := s.security.SelectUserByGUIDAndPass(reqAuth.GUID, reqAuth.Password)
	if err != nil {
		write(w, http.StatusInternalServerError, responseError{Error: http.StatusText(http.StatusInternalServerError)})
		return
	}

	if user.GUID == "" {
		write(w, http.StatusBadRequest, responseError{Error: http.StatusText(http.StatusBadRequest)})
		return
	}

	access, refresh, err := s.security.GenerateTokens(reqAuth.GUID, r.RemoteAddr)
	if err != nil {
		write(w, http.StatusInternalServerError, responseError{Error: http.StatusText(http.StatusInternalServerError)})
		return
	}

	write(w, http.StatusCreated, responseTokens{Access: access, Refresh: refresh})
}

func (s *Server) validateAccessMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization-Access") == "" {
			write(w, http.StatusForbidden, responseError{Error: http.StatusText(http.StatusForbidden)})
			return
		}

		_, err := s.security.ValidateAccess(r.Header.Get("Authorization-Access"))
		if err != nil {
			write(w, http.StatusForbidden, responseError{Error: http.StatusText(http.StatusForbidden)})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// @Summary Обновление токена
// @Description Обновление токена доступа пользователя
// @Produce  json
// @Param Authorization-Access header string true "Access token"
// @Param Authorization-Refresh header string true "Refresh token"
// @Success 202 {object} responseTokens
// @Failure 403 {object} responseError
// @Failure 500 {object} responseError
// @Router /api/refresh [get]
func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	access, refresh, err := s.security.RefreshTokens(r.Header.Get("Authorization-Access"), r.Header.Get("Authorization-Refresh"), r.RemoteAddr)
	if err != nil {
		write(w, http.StatusInternalServerError, responseError{Error: http.StatusText(http.StatusInternalServerError)})
		return
	}

	if access == "" || refresh == "" {
		write(w, http.StatusForbidden, responseError{Error: http.StatusText(http.StatusForbidden)})
		return
	}

	write(w, http.StatusAccepted, responseTokens{Access: access, Refresh: refresh})
}
