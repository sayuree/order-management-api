package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sabina/orders-api/internal/models"
)

// Test SetupRoutes creates router with proper routes
func TestSetupRoutes(t *testing.T) {
	service := &mockOrderService{
		CreateOrderFunc: func(ctx context.Context, order *models.Order) error { return nil },
		ListOrdersFunc: func(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
			return &models.PaginatedOrders{Orders: []models.Order{}, Total: 0, Page: 1, Limit: 10, TotalPages: 1}, nil
		},
	}
	h := NewOrderHandler(service, 10, 100)
	router := SetupRoutes(h)
	
	if router == nil {
		t.Error("expected router to be created")
	}

	// Test POST /api/v1/orders route exists
	req := httptest.NewRequest("POST", "/api/v1/orders", strings.NewReader(`{"customer_id":"test","total_amount":100,"status":"pending"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code == http.StatusNotFound {
		t.Error("POST /api/v1/orders route not found")
	}

	// Test GET /api/v1/orders route exists
	req = httptest.NewRequest("GET", "/api/v1/orders", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	if w.Code == http.StatusNotFound {
		t.Error("GET /api/v1/orders route not found")
	}
}

// Test CORS middleware sets proper headers
func TestCorsMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := corsMiddleware(handler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	wrapped.ServeHTTP(w, req)
	
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS Allow-Origin header to be set")
	}
	
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected CORS Allow-Methods header to be set")
	}
	
	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("expected CORS Allow-Headers header to be set")
	}
}

// Test CORS middleware handles OPTIONS requests
func TestCorsMiddleware_OptionsRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for OPTIONS request")
	})

	wrapped := corsMiddleware(handler)
	
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	
	wrapped.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for OPTIONS request, got %d", w.Code)
	}
}

// Test logging middleware calls next handler
func TestLoggingMiddleware(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	wrapped := loggingMiddleware(handler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	wrapped.ServeHTTP(w, req)
	
	if !handlerCalled {
		t.Error("expected handler to be called")
	}
	
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
