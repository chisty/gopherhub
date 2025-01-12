package main

import (
	"time"

	"github.com/chisty/gopherhub/internal/db"
	"github.com/chisty/gopherhub/internal/env"
	"github.com/chisty/gopherhub/internal/mailer"
	"github.com/chisty/gopherhub/internal/store"
	"go.uber.org/zap"
)

//	@title			GopherHub API
//	@version		1.0
//	@description	API for GopherHub, a social network for gophers.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath					/v1
//
//	@securityDefinitions.apiKey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description				JWT Authorization header

const version = "0.0.1"

func main() {
	cfg := config{
		addr:        env.GetString("ADDR", ":8080"),
		apiURL:      env.GetString("DOCS_URL", "localhost:8080"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:4000"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://postgres:postgres@localhost:5432/gopherhub?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 20),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 20),
			maxIdleTime:  env.GetDuration("DB_MAX_IDLE_TIME", 10*time.Minute),
		},
		env:     env.GetString("ENV", "development"),
		version: env.GetString("VERSION", version),
		mail: mailConfig{
			sendGridCfg: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
			expiry:    env.GetDuration("MAIL_EXPIRY", 3*24*time.Hour),
			fromEmail: env.GetString("FROM_EMAIL", ""),
		},
	}

	// Logger

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database connection
	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Panic(err)
	}

	defer db.Close()
	logger.Info("Database connection established")

	storage := store.NewStorage(db)

	mailer := mailer.NewSendGridMailer(cfg.mail.fromEmail, cfg.mail.sendGridCfg.apiKey)

	app := app{
		config: cfg,
		store:  storage,
		logger: logger,
		mailer: mailer,
	}

	mux := app.mux()
	logger.Fatal(app.run(mux))
}
