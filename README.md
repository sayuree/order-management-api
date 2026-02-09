# Orders Management API

A RESTful API for managing orders built with Go, featuring pagination and filtering capabilities.

## Features

- Create, Read, Update, Delete (CRUD) operations for orders
- Pagination support
- Filtering by status, customer, date range
- PostgreSQL database
- Clean architecture with repository pattern

## Project Structure

```
.
├── cmd/
│   └── api/          # Application entry point
├── internal/
│   ├── config/       # Configuration management
│   ├── database/     # Database connection and migrations
│   ├── handlers/     # HTTP handlers
│   ├── models/       # Domain models
│   ├── repository/   # Data access layer
│   └── service/      # Business logic
├── migrations/       # Database migrations
└── pkg/
    └── response/     # Common response utilities
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 14 or higher

## Setup

1. Install dependencies:

```bash
go mod download
```

2. Set up environment variables:

```bash
cp .env.example .env
# Edit .env with your database credentials
```

3. Run database migrations:

```bash
go run cmd/api/main.go migrate
```

4. (Optional) Seed the database with sample data:

```bash
go run cmd/api/main.go seed
```

This will create 50 sample orders with random data for testing purposes.

5. Start the server:

```bash
go run cmd/api/main.go
```

## API Endpoints

### Orders

- `GET /api/v1/orders` - List orders with pagination and filtering
  - Query params: `page`, `limit`, `status`, `customer_id`, `min_amount`, `max_amount`, `from_date`, `to_date`
- `POST /api/v1/orders` - Create new order

## CLI Commands

- `go run cmd/api/main.go` - Start the API server
- `go run cmd/api/main.go migrate` - Run database migrations
- `go run cmd/api/main.go migrate-down` - Rollback database migrations
- `go run cmd/api/main.go seed` - Seed database with 50 sample orders

## Example Requests

### List orders with pagination

```bash
curl "http://localhost:8080/api/v1/orders?page=1&limit=10&status=pending"
```

### Create order

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "123",
    "total_amount": 99.99,
    "status": "pending",
    "items": [
      {"product_id": "prod1", "quantity": 2, "price": 49.99}
    ]
  }'
```
