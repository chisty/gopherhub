package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chisty/gopherhub/internal/auth"
	"github.com/chisty/gopherhub/internal/store"
	"github.com/chisty/gopherhub/internal/store/cache"
	"go.uber.org/zap"
)

func newTestApp(t *testing.T) *app {
	t.Helper()

	logger := zap.NewNop().Sugar()
	mockStore := store.NewMockStore()
	mockCacheStore := cache.NewMockStore()
	testAuth := auth.NewMockAuthenticator()

	return &app{
		logger:        logger,
		store:         mockStore,
		cacheStore:    mockCacheStore,
		authenticator: testAuth,
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	t.Helper()

	if expected != actual {
		t.Errorf("expected status %d, got %d", expected, actual)
	}
}
