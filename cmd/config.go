package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"rest-authentication/config"
)

// Config represents the configuration
type Config struct {
	DB       config.DB
	Server   config.Server
	Security config.Security
	Email    config.Email
}

// getConfig gets the configuration
func getConfig(cmd *cobra.Command) (*Config, error) {
	if err := configureLogger(cmd); err != nil {
		return nil, fmt.Errorf("set loggers: %v", err)
	}

	val := validator.New()

	server, err := getServerConfig(val)
	if err != nil {
		return nil, fmt.Errorf("get server config: %v", err)
	}

	db, err := getDBConfig(val)
	if err != nil {
		return nil, fmt.Errorf("get db config: %v", err)
	}

	security, err := getSecurityConfig(val)
	if err != nil {
		return nil, fmt.Errorf("get security config: %v", err)
	}

	email, err := getEmailConfig(val)
	if err != nil {
		return nil, fmt.Errorf("get email config: %v", err)
	}

	return &Config{
		Server:   server,
		DB:       db,
		Security: security,
		Email:    email,
	}, nil
}

// getServerConfig gets the server configuration
func getServerConfig(val *validator.Validate) (config.Server, error) {
	cfg := config.Server{
		Address: os.Getenv("SERVER_ADDRESS"),
	}

	if err := val.Struct(cfg); err != nil {
		return config.Server{}, fmt.Errorf("validate server config: %v", err)
	}

	return cfg, nil
}

// getDBConfig gets the database configuration
func getDBConfig(val *validator.Validate) (config.DB, error) {
	cfg := config.DB{
		Addr: os.Getenv("DATABASE_ADDRESS"),
	}

	if err := val.Struct(cfg); err != nil {
		return config.DB{}, fmt.Errorf("validate db config: %v", err)
	}

	return cfg, nil
}

// getSecurityConfig gets the security configuration
func getSecurityConfig(val *validator.Validate) (config.Security, error) {
	cfg := config.Security{
		SecretKey: os.Getenv("SECRET_KEY"),
	}

	if err := val.Struct(cfg); err != nil {
		return config.Security{}, fmt.Errorf("validate security config: %v", err)
	}

	return cfg, nil
}

// getEmailConfig gets the email configuration
func getEmailConfig(val *validator.Validate) (config.Email, error) {
	cfg := config.Email{
		Host:     os.Getenv("EMAIL_HOST"),
		Port:     os.Getenv("EMAIL_PORT"),
		Username: os.Getenv("EMAIL_USERNAME"),
		Password: os.Getenv("EMAIL_PASSWORD"),
	}

	if err := val.Struct(cfg); err != nil {
		return config.Email{}, fmt.Errorf("validate email config: %v", err)
	}

	return cfg, nil
}

// configureLogger configures the logger
func configureLogger(cmd *cobra.Command) error {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			return function, fmt.Sprintf(" %s:%d", frame.File, frame.Line)
		},
	})

	log.SetReportCaller(true)

	logLevel, err := cmd.Flags().GetString("log")
	if err != nil {
		return fmt.Errorf("get log flag: %v", err)
	}
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Errorf("parse log level: %v", err)
		log.Error("valid log levels: trace, debug, info, warn, warning, error, fatal, panic")
		level = log.TraceLevel
	}

	log.SetLevel(level)
	log.Tracef("log level: %s", level.String())

	return nil
}

// persistentFlags adds the persistent flags
func persistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("log", "trace", "error")
}
