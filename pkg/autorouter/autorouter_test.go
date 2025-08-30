package autorouter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHandler is a sample handler for testing
type TestHandler struct {
	name string
}

// Create is a valid handler method
func (h *TestHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Created by %s", h.name)
}

// Get is a valid handler method
func (h *TestHandler) Get(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Got data from %s", h.name)
}

// Update is a valid handler method
func (h *TestHandler) Update(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Updated by %s", h.name)
}

// Delete is a valid handler method
func (h *TestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Deleted by %s", h.name)
}

// HandleSomething should be skipped (starts with Handle)
func (h *TestHandler) HandleSomething(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This should not be auto-registered")
}

// invalidMethod has wrong signature - should be skipped
func (h *TestHandler) invalidMethod(w http.ResponseWriter) {
	fmt.Fprintf(w, "Invalid signature")
}

// unexportedMethod is not exported - should be skipped
func (h *TestHandler) unexportedMethod(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Unexported method")
}

// NotAHandlerMethod returns something - should be skipped
func (h *TestHandler) NotAHandlerMethod() string {
	return "not a handler"
}

// TestBasicRegistration tests basic handler registration
func TestBasicRegistration(t *testing.T) {
	mux := http.NewServeMux()
	handler := &TestHandler{name: "test"}

	router := NewAutoRouter(mux, RegistrationOptions{
		Prefix:       "/api/v1/",
		MethodPrefix: "test.",
	})

	err := router.RegisterHandlers(handler)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Test Create endpoint
	req := httptest.NewRequest("POST", "/api/v1/test.Create", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Created by test") {
		t.Errorf("Unexpected response body: %s", w.Body.String())
	}
}

// TestQuickRegister tests the convenience function
func TestQuickRegister(t *testing.T) {
	mux := http.NewServeMux()
	handler := &TestHandler{name: "quick"}

	err := QuickRegister(mux, "/api/v1/", "user.", handler)
	if err != nil {
		t.Fatalf("Quick registration failed: %v", err)
	}

	// Test Get endpoint
	req := httptest.NewRequest("GET", "/api/v1/user.Get", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Got data from quick") {
		t.Errorf("Unexpected response body: %s", w.Body.String())
	}
}

// TestMiddleware tests middleware application
func TestMiddleware(t *testing.T) {
	mux := http.NewServeMux()
	handler := &TestHandler{name: "middleware"}

	// Create a test middleware that adds a header
	testMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test-Middleware", "applied")
			next.ServeHTTP(w, r)
		})
	}

	router := NewAutoRouter(mux, RegistrationOptions{
		Prefix:       "/api/v1/",
		MethodPrefix: "test.",
		Middleware:   []Middleware{testMiddleware},
	})

	err := router.RegisterHandlers(handler)
	if err != nil {
		t.Fatalf("Registration with middleware failed: %v", err)
	}

	// Test that middleware is applied
	req := httptest.NewRequest("GET", "/api/v1/test.Get", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Header().Get("X-Test-Middleware") != "applied" {
		t.Errorf("Middleware was not applied")
	}
}

// TestRegistrationWithAuth tests auth middleware registration
func TestRegistrationWithAuth(t *testing.T) {
	mux := http.NewServeMux()
	handler := &TestHandler{name: "auth"}

	// Create a test auth middleware
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simple auth check
			if r.Header.Get("Authorization") != "Bearer test-token" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	router := NewAutoRouter(mux, RegistrationOptions{
		Prefix:       "/api/v1/",
		MethodPrefix: "auth.",
	})

	err := router.RegisterHandlersWithAuth(handler, authMiddleware)
	if err != nil {
		t.Fatalf("Registration with auth failed: %v", err)
	}

	// Test without auth - should fail
	req := httptest.NewRequest("GET", "/api/v1/auth.Get", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	// Test with auth - should succeed
	req = httptest.NewRequest("GET", "/api/v1/auth.Get", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 with auth, got %d", w.Code)
	}
}

// TestGetRegisteredHandlers tests handler information retrieval
func TestGetRegisteredHandlers(t *testing.T) {
	mux := http.NewServeMux()
	handler := &TestHandler{name: "info"}

	router := NewAutoRouter(mux, RegistrationOptions{
		Prefix:       "/api/v1/",
		MethodPrefix: "test.",
	})

	handlers := router.GetRegisteredHandlers(handler)

	expectedHandlers := map[string]bool{
		"Create": true,
		"Get":    true,
		"Update": true,
		"Delete": true,
	}
	
	if len(handlers) != len(expectedHandlers) {
		t.Errorf("Expected %d handlers, got %d", len(expectedHandlers), len(handlers))
	}

	// Check that all expected handlers are present
	for _, h := range handlers {
		if !expectedHandlers[h.MethodName] {
			t.Errorf("Unexpected method: %s", h.MethodName)
		}
		
		expectedPath := "/api/v1/test." + h.MethodName
		if h.URLPath != expectedPath {
			t.Errorf("Expected path %s for method %s, got %s", expectedPath, h.MethodName, h.URLPath)
		}
		
		// Remove from expected map
		delete(expectedHandlers, h.MethodName)
	}
	
	// Check if any expected handlers are missing
	for missing := range expectedHandlers {
		t.Errorf("Missing expected handler: %s", missing)
	}
}

// TestSingleMethodRegistration tests registering a single method with custom path
func TestSingleMethodRegistration(t *testing.T) {
	mux := http.NewServeMux()
	handler := &TestHandler{name: "single"}

	router := NewAutoRouter(mux, RegistrationOptions{
		Prefix: "/api/v1/",
	})

	err := router.RegisterSingleMethod(handler, "Get", "custom/path")
	if err != nil {
		t.Fatalf("Single method registration failed: %v", err)
	}

	// Test the custom path
	req := httptest.NewRequest("GET", "/api/v1/custom/path", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Got data from single") {
		t.Errorf("Unexpected response body: %s", w.Body.String())
	}
}