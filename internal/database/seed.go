package database

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

var statuses = []string{"pending", "processing", "shipped", "delivered", "cancelled"}

func SeedOrders(db *sql.DB, n int) error {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < n; i++ {
		customerID := fmt.Sprintf("cust-%d", rand.Intn(10)+1)
		totalAmount := rand.Float64()*1000 + 10
		status := statuses[rand.Intn(len(statuses))]
		createdAt := time.Now().Add(-time.Duration(rand.Intn(365)) * 24 * time.Hour)
		updatedAt := createdAt.Add(time.Duration(rand.Intn(10)) * time.Hour)

		var orderID int64
		err := db.QueryRow(
			`INSERT INTO orders (customer_id, total_amount, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			customerID, totalAmount, status, createdAt, updatedAt,
		).Scan(&orderID)
		if err != nil {
			return fmt.Errorf("failed to insert order: %w", err)
		}

		// Add 1-5 items per order
		itemCount := rand.Intn(5) + 1
		for j := 0; j < itemCount; j++ {
			productID := fmt.Sprintf("prod-%d", rand.Intn(20)+1)
			quantity := rand.Intn(5) + 1
			price := rand.Float64()*200 + 5
			_, err := db.Exec(
				`INSERT INTO order_items (order_id, product_id, quantity, price) VALUES ($1, $2, $3, $4)`,
				orderID, productID, quantity, price,
			)
			if err != nil {
				return fmt.Errorf("failed to insert order item: %w", err)
			}
		}
	}
	return nil
}
