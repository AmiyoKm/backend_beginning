package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AmiyoKm/go-backend/internal/auth"
	"github.com/AmiyoKm/go-backend/internal/store"
	"github.com/AmiyoKm/go-backend/internal/store/cache"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T, cfg Config) *Application {
	t.Helper()
	logger := zap.NewNop().Sugar()
	mockStore := store.NewMockStore()
	mockCacheStore := cache.NewMockStore()
	testAuth := &auth.TestAuthenticator{}
	return &Application{
		Logger:        logger,
		Store:         mockStore,
		CacheStorage:  mockCacheStore,
		Config:        cfg,
		Authenticator: testAuth,
	}
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d", expected, actual)
	}
}
