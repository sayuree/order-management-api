package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/sabina/orders-api/internal/models"
	"github.com/sabina/orders-api/internal/service"
	"github.com/sabina/orders-api/pkg/response"
)

type OrderHandler struct {
	service         service.OrderServiceInterface
	maxPageSize     int
	defaultPageSize int
}

func NewOrderHandler(service service.OrderServiceInterface, defaultPageSize, maxPageSize int) *OrderHandler {
	return &OrderHandler{
		service:       service,
		defaultPageSize: defaultPageSize,
		maxPageSize:   maxPageSize,
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.CreateOrder(r.Context(), &order); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = h.defaultPageSize
	}
	if limit > h.maxPageSize {
		limit = h.maxPageSize
	}

	pagination := &models.Pagination{
		Page:  page,
		Limit: limit,
	}

	// Parse filters
	filter := &models.OrderFilter{}
	
	if customerID := r.URL.Query().Get("customer_id"); customerID != "" {
		filter.CustomerID = &customerID
	}
	
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}
	
	if minAmountStr := r.URL.Query().Get("min_amount"); minAmountStr != "" {
		if minAmount, err := strconv.ParseFloat(minAmountStr, 64); err == nil {
			filter.MinAmount = &minAmount
		}
	}
	
	if maxAmountStr := r.URL.Query().Get("max_amount"); maxAmountStr != "" {
		if maxAmount, err := strconv.ParseFloat(maxAmountStr, 64); err == nil {
			filter.MaxAmount = &maxAmount
		}
	}
	
	if fromDateStr := r.URL.Query().Get("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			filter.FromDate = &fromDate
		}
	}
	
	if toDateStr := r.URL.Query().Get("to_date"); toDateStr != "" {
		if toDate, err := time.Parse("2006-01-02", toDateStr); err == nil {
			filter.ToDate = &toDate
		}
	}

	result, err := h.service.ListOrders(r.Context(), filter, pagination)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, result)
}

