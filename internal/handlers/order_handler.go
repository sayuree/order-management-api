package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
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
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err != nil || parsedPage < 1 {
			response.Error(w, http.StatusBadRequest, "invalid page")
			return
		}
		page = parsedPage
	}

	limit := h.defaultPageSize
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 {
			response.Error(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = parsedLimit
	}
	if limit > h.maxPageSize {
		limit = h.maxPageSize
	}

	pagination := &models.Pagination{
		Page:  page,
		Limit: limit,
		Offset: (page - 1) * limit,
	}

	// Parse filters
	filter := &models.OrderFilter{}
	
	if customerID := r.URL.Query().Get("customer_id"); customerID != "" {
		filter.CustomerID = &customerID
	}
	
	if status := r.URL.Query().Get("status"); status != "" {
		if !models.OrderStatus(status).IsValid() {
			response.Error(w, http.StatusBadRequest, "invalid status")
			return
		}
		filter.Status = &status
	}

	if amountRange := r.URL.Query().Get("amount"); amountRange != "" {
		parts := strings.Split(amountRange, ",")
		if len(parts) > 2 {
			response.Error(w, http.StatusBadRequest, "invalid amount")
			return
		}
		if len(parts) == 1 {
			valueStr := strings.TrimSpace(parts[0])
			if valueStr == "" {
				response.Error(w, http.StatusBadRequest, "invalid amount")
				return
			}
			amount, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				response.Error(w, http.StatusBadRequest, "invalid amount")
				return
			}
			filter.MinAmount = &amount
			filter.MaxAmount = &amount
		} else {
			minStr := strings.TrimSpace(parts[0])
			maxStr := strings.TrimSpace(parts[1])
			if minStr == "" && maxStr == "" {
				response.Error(w, http.StatusBadRequest, "invalid amount")
				return
			}
			if minStr != "" {
				minAmount, err := strconv.ParseFloat(minStr, 64)
				if err != nil {
					response.Error(w, http.StatusBadRequest, "invalid amount")
					return
				}
				filter.MinAmount = &minAmount
			}
			if maxStr != "" {
				maxAmount, err := strconv.ParseFloat(maxStr, 64)
				if err != nil {
					response.Error(w, http.StatusBadRequest, "invalid amount")
					return
				}
				filter.MaxAmount = &maxAmount
			}
			if filter.MinAmount != nil && filter.MaxAmount != nil && *filter.MinAmount > *filter.MaxAmount {
				response.Error(w, http.StatusBadRequest, "invalid amount range")
				return
			}
		}
	}

	if dateRange := r.URL.Query().Get("dateRange"); dateRange != "" {
		parts := strings.Split(dateRange, ",")
		if len(parts) > 2 {
			response.Error(w, http.StatusBadRequest, "invalid dateRange")
			return
		}
		fromStr := ""
		toStr := ""
		if len(parts) > 0 {
			fromStr = strings.TrimSpace(parts[0])
		}
		if len(parts) > 1 {
			toStr = strings.TrimSpace(parts[1])
		}
		if fromStr == "" && toStr == "" {
			response.Error(w, http.StatusBadRequest, "invalid dateRange")
			return
		}
		if fromStr != "" {
			fromDate, err := time.Parse("2006-01-02", fromStr)
			if err != nil {
				response.Error(w, http.StatusBadRequest, "invalid dateRange")
				return
			}
			filter.FromDate = &fromDate
		}
		if toStr != "" {
			toDate, err := time.Parse("2006-01-02", toStr)
			if err != nil {
				response.Error(w, http.StatusBadRequest, "invalid dateRange")
				return
			}
			filter.ToDate = &toDate
		}
		if filter.FromDate != nil && filter.ToDate != nil && filter.FromDate.After(*filter.ToDate) {
			response.Error(w, http.StatusBadRequest, "invalid dateRange")
			return
		}
	}
	
	if filter.MinAmount == nil && filter.MaxAmount == nil {
		if minAmountStr := r.URL.Query().Get("min_amount"); minAmountStr != "" {
			minAmount, err := strconv.ParseFloat(minAmountStr, 64)
			if err != nil {
				response.Error(w, http.StatusBadRequest, "invalid min_amount")
				return
			}
			filter.MinAmount = &minAmount
		}
		
		if maxAmountStr := r.URL.Query().Get("max_amount"); maxAmountStr != "" {
			maxAmount, err := strconv.ParseFloat(maxAmountStr, 64)
			if err != nil {
				response.Error(w, http.StatusBadRequest, "invalid max_amount")
				return
			}
			filter.MaxAmount = &maxAmount
		}
		if filter.MinAmount != nil && filter.MaxAmount != nil && *filter.MinAmount > *filter.MaxAmount {
			response.Error(w, http.StatusBadRequest, "invalid amount range")
			return
		}
	}
	
	if filter.FromDate == nil && filter.ToDate == nil {
		if fromDateStr := r.URL.Query().Get("from_date"); fromDateStr != "" {
			fromDate, err := time.Parse("2006-01-02", fromDateStr)
			if err != nil {
				response.Error(w, http.StatusBadRequest, "invalid from_date")
				return
			}
			filter.FromDate = &fromDate
		}
		
		if toDateStr := r.URL.Query().Get("to_date"); toDateStr != "" {
			toDate, err := time.Parse("2006-01-02", toDateStr)
			if err != nil {
				response.Error(w, http.StatusBadRequest, "invalid to_date")
				return
			}
			filter.ToDate = &toDate
		}
		if filter.FromDate != nil && filter.ToDate != nil && filter.FromDate.After(*filter.ToDate) {
			response.Error(w, http.StatusBadRequest, "invalid date range")
			return
		}
	}

	result, err := h.service.ListOrders(r.Context(), filter, pagination)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, result)
}

