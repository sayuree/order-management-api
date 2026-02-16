package models

import (
	"time"
)

type Order struct {
	ID           int64      `json:"id"`
	CustomerID   string     `json:"customer_id"`
	TotalAmount  float64    `json:"total_amount"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Items        []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusProcessing OrderStatus = "processing"
	StatusShipped    OrderStatus = "shipped"
	StatusDelivered  OrderStatus = "delivered"
	StatusCancelled  OrderStatus = "cancelled"
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusProcessing, StatusShipped, StatusDelivered, StatusCancelled:
		return true
	}
	return false
}

type OrderFilter struct {
	CustomerID *string
	Status     *string
	FromDate   *time.Time
	ToDate     *time.Time
	MinAmount  *float64
	MaxAmount  *float64
}

type Pagination struct {
	Page  int
	Limit int
	Offset int
}

type PaginatedOrders struct {
	Orders     []Order `json:"orders"`
	Total      int64   `json:"total"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	Offset     int     `json:"offset"`
	TotalPages int     `json:"total_pages"`
}
