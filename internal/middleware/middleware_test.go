package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogging(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Create the logging middleware
	loggingMiddleware := Logging(logger, nil)

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute the middleware
	loggingMiddleware(testHandler).ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "test response" {
		t.Errorf("expected body 'test response', got '%s'", w.Body.String())
	}

	// Check log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, `"method":"GET"`) {
		t.Errorf("log should contain method GET, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, `"url":"/test"`) {
		t.Errorf("log should contain url /test, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, `"status":200`) {
		t.Errorf("log should contain status 200, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, `"duration"`) {
		t.Errorf("log should contain duration, got: %s", logOutput)
	}
}

func TestRecovery(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Create the recovery middleware
	recoveryMiddleware := Recovery(logger)

	// Create a test request
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// Execute the middleware
	recoveryMiddleware(panicHandler).ServeHTTP(w, req)

	// Check response - should be 500 Internal Server Error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
	expectedBody := "Internal Server Error"
	actualBody := strings.TrimSpace(w.Body.String())
	if actualBody != expectedBody {
		t.Errorf("expected body '%s', got '%s'", expectedBody, actualBody)
	}

	// Check log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, `"panic":"test panic"`) {
		t.Errorf("log should contain panic message, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, `"url":"/panic"`) {
		t.Errorf("log should contain url /panic, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, `"method":"GET"`) {
		t.Errorf("log should contain method GET, got: %s", logOutput)
	}
}

func TestCORS(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Create the CORS middleware
	corsMiddleware := CORS()

	// Test normal GET request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	corsMiddleware(testHandler).ServeHTTP(w, req)

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin '*', got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("expected Access-Control-Allow-Methods 'GET, POST, PUT, DELETE, OPTIONS', got '%s'", w.Header().Get("Access-Control-Allow-Methods"))
	}
	if w.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
		t.Errorf("expected Access-Control-Allow-Headers 'Content-Type, Authorization', got '%s'", w.Header().Get("Access-Control-Allow-Headers"))
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "test response" {
		t.Errorf("expected body 'test response', got '%s'", w.Body.String())
	}

	// Test OPTIONS preflight request
	optionsReq := httptest.NewRequest("OPTIONS", "/test", nil)
	optionsW := httptest.NewRecorder()

	corsMiddleware(testHandler).ServeHTTP(optionsW, optionsReq)

	// Check OPTIONS response
	if optionsW.Code != http.StatusOK {
		t.Errorf("expected OPTIONS status 200, got %d", optionsW.Code)
	}
	if optionsW.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected OPTIONS Access-Control-Allow-Origin '*', got '%s'", optionsW.Header().Get("Access-Control-Allow-Origin"))
	}
	if optionsW.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("expected OPTIONS Access-Control-Allow-Methods 'GET, POST, PUT, DELETE, OPTIONS', got '%s'", optionsW.Header().Get("Access-Control-Allow-Methods"))
	}
	if optionsW.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
		t.Errorf("expected OPTIONS Access-Control-Allow-Headers 'Content-Type, Authorization', got '%s'", optionsW.Header().Get("Access-Control-Allow-Headers"))
	}
	if optionsW.Body.String() != "" {
		t.Errorf("expected OPTIONS empty body, got '%s'", optionsW.Body.String())
	}
}
