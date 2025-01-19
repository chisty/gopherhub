package main

import (
	"time"

	"github.com/chisty/gopherhub/internal/auth"
	"github.com/chisty/gopherhub/internal/db"
	"github.com/chisty/gopherhub/internal/env"
	"github.com/chisty/gopherhub/internal/mailer"
	"github.com/chisty/gopherhub/internal/store"
	"github.com/chisty/gopherhub/internal/store/cache"
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
		addr:        env.GetString("ADDR", ":8180"),
		apiURL:      env.GetString("DOCS_URL", "localhost:8180"),
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:5173"),
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
				apiKey: env.GetString("SENDGRID_API_KEY", "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"),
			},
			expiry:    env.GetDuration("MAIL_EXPIRY", 3*24*time.Hour),
			fromEmail: env.GetString("FROM_EMAIL", "chisty.kaz@gmail.com"),
		},
		auth: authConfig{
			basic: basicConfig{
				username: env.GetString("BASIC_AUTH_USERNAME", "admin"),
				password: env.GetString("BASIC_AUTH_PASSWORD", "admin"),
			},
			token: tokenConfig{
				secret:   env.GetString("AUTH_TOKEN_SECRET", "secret"),
				issuer:   env.GetString("AUTH_TOKEN_ISSUER", "gopherhub"),
				audience: env.GetString("AUTH_TOKEN_AUDIENCE", "gopherhub"),
				expiry:   env.GetDuration("AUTH_TOKEN_EXPIRY", 24*time.Hour),
			},
		},
		redisCfg: redisConfig{
			addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
			pw:      env.GetString("REDIS_PW", ""),
			db:      env.GetInt("REDIS_DB", 0),
			enabled: true,
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

	// Cache
	redisClient := cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pw, cfg.redisCfg.db)
	logger.Info("Redis connection established")

	storage := store.NewStorage(db)
	cacheStorage := cache.NewRedisStorage(redisClient)

	mailer, err := mailer.NewSendGridMailer(cfg.mail.fromEmail, cfg.mail.sendGridCfg.apiKey)
	if err != nil {
		panic(err)
	}

	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.audience, cfg.auth.token.issuer)

	app := app{
		config:        cfg,
		store:         storage,
		cacheStore:    cacheStorage,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
	}

	mux := app.mux()
	logger.Fatal(app.run(mux))
}
