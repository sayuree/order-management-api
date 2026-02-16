package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/sabina/orders-api/internal/models"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

// NewPostgresOrderRepository creates a new PostgresOrderRepository
func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

// Create inserts a new order and its items into the database
func (r *PostgresOrderRepository) Create(ctx context.Context, order *models.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO orders (customer_id, total_amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRowContext(ctx, query, order.CustomerID, order.TotalAmount, order.Status).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	if len(order.Items) > 0 {
		itemQuery := `
			INSERT INTO order_items (order_id, product_id, quantity, price)
			VALUES ($1, $2, $3, $4)
		`
		for _, item := range order.Items {
			_, err := tx.ExecContext(ctx, itemQuery, order.ID, item.ProductID, item.Quantity, item.Price)
			if err != nil {
				return fmt.Errorf("failed to insert order item: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
	}

func (r *PostgresOrderRepository) List(ctx context.Context, filter *models.OrderFilter, pagination *models.Pagination) (*models.PaginatedOrders, error) {
	if pagination == nil {
		pagination = &models.Pagination{Page: 1, Limit: 10, Offset: 0}
	}
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.Limit < 1 {
		pagination.Limit = 10
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Build WHERE clause
	if filter != nil {
		if filter.CustomerID != nil {
			conditions = append(conditions, fmt.Sprintf("customer_id = $%d", argIndex))
			args = append(args, *filter.CustomerID)
			argIndex++
		}
		if filter.Status != nil {
			conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, *filter.Status)
			argIndex++
		}
		if filter.FromDate != nil {
			conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
			args = append(args, *filter.FromDate)
			argIndex++
		}
		if filter.ToDate != nil {
			conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
			args = append(args, *filter.ToDate)
			argIndex++
		}
		if filter.MinAmount != nil {
			conditions = append(conditions, fmt.Sprintf("total_amount >= $%d", argIndex))
			args = append(args, *filter.MinAmount)
			argIndex++
		}
		if filter.MaxAmount != nil {
			conditions = append(conditions, fmt.Sprintf("total_amount <= $%d", argIndex))
			args = append(args, *filter.MaxAmount)
			argIndex++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM orders %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count orders: %w", err)
	}

	// Get paginated results
	expectedOffset := (pagination.Page - 1) * pagination.Limit
	offset := pagination.Offset
	if offset < 0 || offset != expectedOffset {
		offset = expectedOffset
	}
	query := fmt.Sprintf(`
		SELECT id, customer_id, total_amount, status, created_at, updated_at
		FROM orders
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, pagination.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.TotalAmount,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate orders: %w", err)
	}

	totalPages := int(total) / pagination.Limit
	if int(total)%pagination.Limit > 0 {
		totalPages++
	}

	return &models.PaginatedOrders{
		Orders:     orders,
		Total:      total,
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		Offset:     offset,
		TotalPages: totalPages,
	}, nil
}


