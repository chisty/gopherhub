package main

import (
	"expvar"
	"fmt"
	"runtime"
	"time"

	"github.com/chisty/gopherhub/internal/auth"
	"github.com/chisty/gopherhub/internal/db"
	"github.com/chisty/gopherhub/internal/env"
	"github.com/chisty/gopherhub/internal/mailer"
	"github.com/chisty/gopherhub/internal/ratelimiter"
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
		ratelimiterCfg: ratelimiter.Config{
			RequestPerTimeFrame: env.GetInt("RATE_LIMITER_REQUEST_PER_TIME_FRAME", 20),
			TimeFrame:           env.GetDuration("RATE_LIMITER_TIME_FRAME", 5*time.Second),
			Enabled:             true,
		},
	}

	// Initialize structured logger based on environment
	var logger *zap.SugaredLogger
	if cfg.env == "development" {
		logger = zap.Must(zap.NewDevelopment()).Sugar()
	} else {
		logger = zap.Must(zap.NewProduction()).Sugar()
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			// Handle sync error gracefully in production
			fmt.Printf("Logger sync error: %v\n", err)
		}
	}()

	// Log application startup information
	logger.Infow("Starting GopherHub API",
		"version", cfg.version,
		"environment", cfg.env,
		"address", cfg.addr,
		"go_version", runtime.Version(),
		"num_cpu", runtime.NumCPU(),
	)

	// Database connection with enhanced error handling
	logger.Info("Establishing database connection...")
	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		logger.Fatalw("Failed to establish database connection",
			"error", err,
			"db_addr", cfg.db.addr,
		)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Errorw("Error closing database connection", "error", err)
		}
	}()
	logger.Infow("Database connection established successfully",
		"max_open_conns", cfg.db.maxOpenConns,
		"max_idle_conns", cfg.db.maxIdleConns,
	)

	// Cache connection with enhanced error handling
	logger.Info("Establishing Redis cache connection...")
	redisClient := cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pw, cfg.redisCfg.db)
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Errorw("Error closing Redis connection", "error", err)
		}
	}()
	logger.Infow("Redis connection established successfully",
		"redis_addr", cfg.redisCfg.addr,
		"redis_db", cfg.redisCfg.db,
	)

	// Initialize components
	ratelimiter := ratelimiter.NewFixedWindowLimiter(cfg.ratelimiterCfg.RequestPerTimeFrame, cfg.ratelimiterCfg.TimeFrame)
	storage := store.NewStorage(db)
	cacheStorage := cache.NewRedisStorage(redisClient)

	// Initialize mailer with error handling
	logger.Info("Initializing mailer service...")
	mailer, err := mailer.NewSendGridMailer(cfg.mail.fromEmail, cfg.mail.sendGridCfg.apiKey)
	if err != nil {
		logger.Fatalw("Failed to initialize mailer service",
			"error", err,
			"from_email", cfg.mail.fromEmail,
		)
	}
	logger.Info("Mailer service initialized successfully")

	// Initialize JWT authenticator
	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.audience, cfg.auth.token.issuer)
	logger.Infow("JWT authenticator initialized",
		"issuer", cfg.auth.token.issuer,
		"audience", cfg.auth.token.audience,
	)

	app := app{
		config:        cfg,
		store:         storage,
		cacheStore:    cacheStorage,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
		rateLimiter:   ratelimiter,
	}

	// Enhanced metrics collection
	logger.Info("Setting up application metrics...")

	// Version and build info
	expvar.NewString("version").Set(version)
	expvar.NewString("environment").Set(cfg.env)
	expvar.NewString("go_version").Set(runtime.Version())

	// Runtime metrics
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("memory", expvar.Func(func() interface{} {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return map[string]interface{}{
			"alloc_mb":      m.Alloc / 1024 / 1024,
			"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
			"sys_mb":        m.Sys / 1024 / 1024,
			"gc_runs":       m.NumGC,
		}
	}))

	// Database connection metrics
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	// Application health status
	expvar.Publish("health", expvar.Func(func() interface{} {
		return map[string]interface{}{
			"status":     "healthy",
			"timestamp":  time.Now().Unix(),
			"uptime":     time.Since(time.Now()).String(), // This would be calculated properly in real implementation
		}
	}))

	logger.Info("Application metrics configured successfully")

	// Start the server
	mux := app.mux()
	logger.Infow("Starting HTTP server", "address", cfg.addr)
	
	if err := app.run(mux); err != nil {
		logger.Fatalw("Server failed to start", "error", err)
	}
}