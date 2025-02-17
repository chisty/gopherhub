package main

import (
	"expvar"
	"fmt"
	"net/http"
	"time"

	"github.com/chisty/gopherhub/docs"
	"github.com/chisty/gopherhub/internal/auth"
	"github.com/chisty/gopherhub/internal/mailer"
	"github.com/chisty/gopherhub/internal/ratelimiter"
	"github.com/chisty/gopherhub/internal/store"
	"github.com/chisty/gopherhub/internal/store/cache"
	"github.com/chisty/gopherhub/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type app struct {
	config        config
	store         store.Storage
	cacheStore    cache.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
	rateLimiter   ratelimiter.Limiter
}

type config struct {
	addr           string
	db             dbConfig
	env            string
	version        string
	apiURL         string
	mail           mailConfig
	frontendURL    string
	auth           authConfig
	redisCfg       redisConfig
	ratelimiterCfg ratelimiter.Config
}

type redisConfig struct {
	addr    string
	pw      string
	db      int
	enabled bool
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type tokenConfig struct {
	secret   string
	issuer   string
	audience string
	expiry   time.Duration
}

type basicConfig struct {
	username string
	password string
}

type mailConfig struct {
	sendGridCfg sendGridConfig
	fromEmail   string
	expiry      time.Duration
}

type sendGridConfig struct {
	apiKey string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  time.Duration
}

func (app *app) mux() http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// should be before rate limiter
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(app.RateLimiterMiddleware)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		// r.With(app.AuthBasicMiddleware()).Get("/health", app.healthCheckHandler)
		r.Get("/health", app.healthCheckHandler)
		r.With(app.AuthBasicMiddleware()).Get("/debug/vars", expvar.Handler().ServeHTTP)

		swaggerURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(swaggerURL),
		))

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)

			r.Post("/", app.createPostHandler)

			r.Route("/{id}", func(r chi.Router) {
				r.Use(app.postContextMiddleware)

				r.Get("/", app.getPostHandler)
				r.Patch("/", app.checkPostOwnership("moderator", app.updatePostHandler))
				r.Delete("/", app.checkPostOwnership("admin", app.deletePostHandler))

				r.Post("/comments", app.createCommentHandler)
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route("/{id}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)

				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
		})
	})

	return r
}

func (app *app) run(mux http.Handler) error {

	// Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute * 15,
	}

	app.logger.Infow("Server has started", "addr", app.config.addr, "env", app.config.env)

	return srv.ListenAndServe()
}

func (app *app) jsonResponse(w http.ResponseWriter, code int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}

	return util.WriteJSON(w, code, &envelope{Data: data})
}
