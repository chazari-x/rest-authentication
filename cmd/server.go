package cmd

import (
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"rest-authentication/email"
	"rest-authentication/security"
	"rest-authentication/server"
	"rest-authentication/storage"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Info("No .env file found.")
	}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "server",
		Long:  "server",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := getConfig(cmd)
			if err != nil {
				log.Fatalf("config err: %s", err)
			}

			log.Trace("server starting..")
			defer log.Trace("server stopped")

			store, err := storage.New(cmd.Context(), cfg.DB)
			if err != nil {
				log.Fatalf("storage err: %s", err)
			}

			defer store.Close()

			mail := email.New(cfg.Email, store)

			secure := security.New(cfg.Security, store, mail)

			if err = server.New(cfg.Server, secure).Start(); err != nil {
				log.Error(err)
			}
		},
	}
	rootCmd.AddCommand(cmd)
	persistentFlags(cmd)
}
