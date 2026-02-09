package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sabina/orders-api/internal/models"
)

type mockOrderService struct {
	CreateOrderFunc func(ctx context.Context, order *models.Order) error
	ListOrdersFunc  func(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error)
}

func (m *mockOrderService) CreateOrder(ctx context.Context, order *models.Order) error {
	return m.CreateOrderFunc(ctx, order)
}
func (m *mockOrderService) ListOrders(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
	return m.ListOrdersFunc(ctx, filter, pagination)
}

func setupTestHandler() *OrderHandler {
	service := &mockOrderService{
		CreateOrderFunc: func(ctx context.Context, order *models.Order) error { return nil },
		ListOrdersFunc: func(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
			return &models.PaginatedOrders{Orders: []models.Order{}, Total: 0, Page: 1, Limit: 10, TotalPages: 1}, nil
		},
	}
	return NewOrderHandler(service, 10, 100)
}

// 1. Test valid order creation
func TestCreateOrder_Valid(t *testing.T) {
	h := setupTestHandler()
	order := models.Order{CustomerID: "cust-1", TotalAmount: 100.50, Status: "pending"}
	body, _ := json.Marshal(order)
	req := httptest.NewRequest("POST", "/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.CreateOrder(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var response models.Order
	json.NewDecoder(w.Body).Decode(&response)
	if response.CustomerID != "cust-1" {
		t.Errorf("expected customer_id cust-1, got %s", response.CustomerID)
	}
}

// 2. Test invalid JSON body
func TestCreateOrder_InvalidBody(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("POST", "/api/v1/orders", bytes.NewReader([]byte("notjson")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.CreateOrder(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// 3. Test service validation error
func TestCreateOrder_ValidationError(t *testing.T) {
	service := &mockOrderService{
		CreateOrderFunc: func(ctx context.Context, order *models.Order) error {
			return errors.New("customer_id is required")
		},
	}
	h := NewOrderHandler(service, 10, 100)
	order := models.Order{TotalAmount: 100, Status: "pending"}
	body, _ := json.Marshal(order)
	req := httptest.NewRequest("POST", "/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.CreateOrder(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// 4. Test default pagination
func TestListOrders_DefaultPagination(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest("GET", "/api/v1/orders", nil)
	w := httptest.NewRecorder()
	h.ListOrders(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// 5. Test custom pagination
func TestListOrders_CustomPagination(t *testing.T) {
	service := &mockOrderService{
		ListOrdersFunc: func(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
			if pagination.Page != 2 || pagination.Limit != 5 {
				t.Errorf("expected page=2, limit=5, got page=%d, limit=%d", pagination.Page, pagination.Limit)
			}
			return &models.PaginatedOrders{Orders: []models.Order{}, Total: 0, Page: 2, Limit: 5, TotalPages: 1}, nil
		},
	}
	h := NewOrderHandler(service, 10, 100)
	req := httptest.NewRequest("GET", "/api/v1/orders?page=2&limit=5", nil)
	w := httptest.NewRecorder()
	h.ListOrders(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// 6. Test pagination exceeds max
func TestListOrders_PaginationExceedsMax(t *testing.T) {
	service := &mockOrderService{
		ListOrdersFunc: func(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
			if pagination.Limit != 100 {
				t.Errorf("expected limit capped at 100, got %d", pagination.Limit)
			}
			return &models.PaginatedOrders{Orders: []models.Order{}, Total: 0, Page: 1, Limit: 100, TotalPages: 1}, nil
		},
	}
	h := NewOrderHandler(service, 10, 100)
	req := httptest.NewRequest("GET", "/api/v1/orders?limit=200", nil)
	w := httptest.NewRecorder()
	h.ListOrders(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// 7. Test status filter
func TestListOrders_StatusFilter(t *testing.T) {
	service := &mockOrderService{
		ListOrdersFunc: func(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
			if filter.Status == nil || *filter.Status != "shipped" {
				t.Error("expected status filter 'shipped'")
			}
			return &models.PaginatedOrders{Orders: []models.Order{}, Total: 0, Page: 1, Limit: 10, TotalPages: 1}, nil
		},
	}
	h := NewOrderHandler(service, 10, 100)
	req := httptest.NewRequest("GET", "/api/v1/orders?status=shipped", nil)
	w := httptest.NewRecorder()
	h.ListOrders(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// 8. Test customer ID filter
func TestListOrders_CustomerIDFilter(t *testing.T) {
	service := &mockOrderService{
		ListOrdersFunc: func(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
			if filter.CustomerID == nil || *filter.CustomerID != "cust-123" {
				t.Error("expected customer_id filter 'cust-123'")
			}
			return &models.PaginatedOrders{Orders: []models.Order{}, Total: 0, Page: 1, Limit: 10, TotalPages: 1}, nil
		},
	}
	h := NewOrderHandler(service, 10, 100)
	req := httptest.NewRequest("GET", "/api/v1/orders?customer_id=cust-123", nil)
	w := httptest.NewRecorder()
	h.ListOrders(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// 9. Test amount range filter
func TestListOrders_AmountRange(t *testing.T) {
	service := &mockOrderService{
		ListOrdersFunc: func(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
			if filter.MinAmount == nil || *filter.MinAmount != 50.0 {
				t.Error("expected min_amount 50.0")
			}
			if filter.MaxAmount == nil || *filter.MaxAmount != 200.0 {
				t.Error("expected max_amount 200.0")
			}
			return &models.PaginatedOrders{Orders: []models.Order{}, Total: 0, Page: 1, Limit: 10, TotalPages: 1}, nil
		},
	}
	h := NewOrderHandler(service, 10, 100)
	req := httptest.NewRequest("GET", "/api/v1/orders?min_amount=50&max_amount=200", nil)
	w := httptest.NewRecorder()
	h.ListOrders(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// 10. Test date range filter
func TestListOrders_DateRange(t *testing.T) {
	from := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	to := time.Now().Format("2006-01-02")
	service := &mockOrderService{
		ListOrdersFunc: func(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
			if filter.FromDate == nil {
				t.Error("expected from_date to be set")
			}
			if filter.ToDate == nil {
				t.Error("expected to_date to be set")
			}
			return &models.PaginatedOrders{Orders: []models.Order{}, Total: 0, Page: 1, Limit: 10, TotalPages: 1}, nil
		},
	}
	h := NewOrderHandler(service, 10, 100)
	req := httptest.NewRequest("GET", "/api/v1/orders?from_date="+from+"&to_date="+to, nil)
	w := httptest.NewRecorder()
	h.ListOrders(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// 11. Test all filters combined
func TestListOrders_AllFilters(t *testing.T) {
	from := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	to := time.Now().Format("2006-01-02")
	service := &mockOrderService{
		ListOrdersFunc: func(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
			if filter.Status == nil || *filter.Status != "delivered" {
				t.Error("expected status 'delivered'")
			}
			if filter.MinAmount == nil || filter.MaxAmount == nil {
				t.Error("expected amount range")
			}
			if filter.FromDate == nil || filter.ToDate == nil {
				t.Error("expected date range")
			}
			return &models.PaginatedOrders{Orders: []models.Order{}, Total: 0, Page: 1, Limit: 10, TotalPages: 1}, nil
		},
	}
	h := NewOrderHandler(service, 10, 100)
	req := httptest.NewRequest("GET", "/api/v1/orders?status=delivered&min_amount=10&max_amount=1000&from_date="+from+"&to_date="+to, nil)
	w := httptest.NewRecorder()
	h.ListOrders(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// 12. Test invalid min_amount (should be ignored)
func TestListOrders_InvalidMinAmount(t *testing.T) {
	service := &mockOrderService{
		ListOrdersFunc: func(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
			if filter.MinAmount != nil {
				t.Error("expected min_amount to be nil when invalid")
			}
			return &models.PaginatedOrders{Orders: []models.Order{}, Total: 0, Page: 1, Limit: 10, TotalPages: 1}, nil
		},
	}
	h := NewOrderHandler(service, 10, 100)
	req := httptest.NewRequest("GET", "/api/v1/orders?min_amount=notanumber", nil)
	w := httptest.NewRecorder()
	h.ListOrders(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// 13. Test invalid date format (should be ignored)
func TestListOrders_InvalidFromDate(t *testing.T) {
	service := &mockOrderService{
		ListOrdersFunc: func(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
			if filter.FromDate != nil {
				t.Error("expected from_date to be nil when invalid")
			}
			return &models.PaginatedOrders{Orders: []models.Order{}, Total: 0, Page: 1, Limit: 10, TotalPages: 1}, nil
		},
	}
	h := NewOrderHandler(service, 10, 100)
	req := httptest.NewRequest("GET", "/api/v1/orders?from_date=bad-date", nil)
	w := httptest.NewRecorder()
	h.ListOrders(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// 14. Test service error on list
func TestListOrders_ServiceError(t *testing.T) {
	service := &mockOrderService{
		ListOrdersFunc: func(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
			return nil, errors.New("database connection failed")
		},
	}
	h := NewOrderHandler(service, 10, 100)
	req := httptest.NewRequest("GET", "/api/v1/orders", nil)
	w := httptest.NewRecorder()
	h.ListOrders(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// 15. Test zero/negative pagination (should use defaults)
func TestListOrders_NegativePagination(t *testing.T) {
	service := &mockOrderService{
		ListOrdersFunc: func(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
			if pagination.Page != 1 {
				t.Errorf("expected page to default to 1, got %d", pagination.Page)
			}
			if pagination.Limit != 10 {
				t.Errorf("expected limit to default to 10, got %d", pagination.Limit)
			}
			return &models.PaginatedOrders{Orders: []models.Order{}, Total: 0, Page: 1, Limit: 10, TotalPages: 1}, nil
		},
	}
	h := NewOrderHandler(service, 10, 100)
	req := httptest.NewRequest("GET", "/api/v1/orders?page=-1&limit=0", nil)
	w := httptest.NewRecorder()
	h.ListOrders(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
