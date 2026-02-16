package service

import (
	"context"
	"fmt"

	"github.com/sabina/orders-api/internal/models"
	"github.com/sabina/orders-api/internal/repository"
)

// OrderServiceInterface defines the contract for order service
type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	ListOrders(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error)
}

type OrderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

func (s *OrderService) CreateOrder(ctx context.Context, order *models.Order) error {
	if err := s.validateOrder(order); err != nil {
		return err
	}
	return s.repo.Create(ctx, order)
}


func (s *OrderService) ListOrders(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
	   if pagination == nil {
	       pagination = &models.Pagination{Page: 1, Limit: 10, Offset: 0}
	   }
       // Validate pagination
       if pagination.Page < 1 {
	       pagination.Page = 1
       }
       if pagination.Limit < 1 {
	       pagination.Limit = 10
       }
	   if pagination.Offset < 0 || pagination.Offset != (pagination.Page-1)*pagination.Limit {
	       pagination.Offset = (pagination.Page - 1) * pagination.Limit
	   }
       return s.repo.List(ctx, filter, pagination)
}

func (s *OrderService) validateOrder(order *models.Order) error {
       if order.CustomerID == "" {
	       return fmt.Errorf("customer_id is required")
       }
       if order.TotalAmount < 0 {
	       return fmt.Errorf("total_amount must be non-negative")
       }
       if order.Status == "" {
	       order.Status = string(models.StatusPending)
       }
       if !models.OrderStatus(order.Status).IsValid() {
	       return fmt.Errorf("invalid order status: %s", order.Status)
       }
       return nil
}
