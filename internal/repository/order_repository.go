package repository

import (
	"context"

	"github.com/sabina/orders-api/internal/models"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	List(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error)
}
