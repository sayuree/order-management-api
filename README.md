# Orders Management API

A RESTful API for managing orders built with Go, featuring pagination, filtering, and clean architecture. This API provides endpoints for creating and retrieving orders with comprehensive filtering options including status, customer ID, amount range, and date range.

## Table of Contents

- [Features](#features)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Setup Guide](#setup-guide)
- [API Endpoints](#api-endpoints)
- [Usage Examples](#usage-examples)
- [Testing](#testing)
- [CLI Commands](#cli-commands)

## Features

- **Order Management**: Create and retrieve orders with associated items
- **Pagination**: Configurable page size with limits
- **Advanced Filtering**: Filter by status, customer ID, amount range, and date range
- **PostgreSQL Database**: Robust relational database with migrations
- **Clean Architecture**: Repository pattern with clear separation of concerns
- **Auto Database Creation**: Automatically creates database if it doesn't exist
- **Sample Data Seeding**: Built-in command to generate test data
- **Test Coverage**: 100% code coverage for handlers with 19 comprehensive tests
- **CORS Support**: Cross-origin resource sharing enabled
- **Logging Middleware**: Request logging for debugging

## Project Structure

```
.
├── cmd/
│   └── api/              # Application entry point
│       └── main.go       # Main application with CLI commands
├── internal/
│   ├── config/           # Configuration management
│   │   └── config.go     # Environment variable loading
│   ├── database/         # Database operations
│   │   ├── database.go   # Connection and auto-creation
│   │   ├── migrations.go # Migration runner
│   │   └── seed.go       # Sample data seeder
│   ├── handlers/         # HTTP handlers
│   │   ├── order_handler.go      # Order endpoints
│   │   ├── order_handler_test.go # Handler tests
│   │   ├── routes.go             # Route setup
│   │   └── routes_test.go        # Route tests
│   ├── models/           # Domain models
│   │   └── order.go      # Order, OrderItem, filters
│   ├── repository/       # Data access layer
│   │   ├── order_repository.go          # Repository interface
│   │   └── postgres_order_repository.go # PostgreSQL implementation
│   └── service/          # Business logic
│       └── order_service.go # Order service with validation
├── migrations/           # Database migrations
│   ├── 000001_create_orders_tables.up.sql
│   └── 000001_create_orders_tables.down.sql
├── pkg/
│   └── response/         # Common response utilities
│       └── response.go   # JSON response helpers
├── .env                  # Environment configuration
├── .env.example          # Example environment file
├── docker-compose.yml    # Docker setup for PostgreSQL
├── Makefile              # Build and run commands
└── README.md             # This file
```

## Prerequisites

- **Go**: Version 1.21 or higher ([Download](https://golang.org/dl/))
- **PostgreSQL**: Version 14 or higher ([Download](https://www.postgresql.org/download/))
- **Git**: For cloning the repository

## Installation

1. **Clone the repository**:

```bash
git clone <repository-url>
cd homework
```

2. **Install Go dependencies**:

```bash
go mod download
```

3. **Verify installation**:

```bash
go version  # Should show Go 1.21+
psql --version  # Should show PostgreSQL 14+
```

## Configuration

The application uses environment variables for configuration. Create a `.env` file in the project root:

```bash
cp .env.example .env
```

Edit the `.env` file with your settings:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=orders_db
DB_SSLMODE=disable

# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080

# Pagination Configuration
DEFAULT_PAGE_SIZE=10
MAX_PAGE_SIZE=100
```

### Configuration Options

| Variable            | Description                | Default     |
| ------------------- | -------------------------- | ----------- |
| `DB_HOST`           | PostgreSQL host            | `localhost` |
| `DB_PORT`           | PostgreSQL port            | `5432`      |
| `DB_USER`           | Database user              | `postgres`  |
| `DB_PASSWORD`       | Database password          | `postgres`  |
| `DB_NAME`           | Database name              | `orders_db` |
| `DB_SSLMODE`        | SSL mode (disable/require) | `disable`   |
| `SERVER_HOST`       | API server host            | `localhost` |
| `SERVER_PORT`       | API server port            | `8080`      |
| `DEFAULT_PAGE_SIZE` | Default pagination size    | `10`        |
| `MAX_PAGE_SIZE`     | Maximum pagination size    | `100`       |

## Setup Guide

### Step 1: Start PostgreSQL

**Option A: Using Docker** (Recommended):

```bash
docker-compose up -d
```

**Option B: Using local PostgreSQL**:

```bash
# macOS (Homebrew)
brew services start postgresql

# Linux
sudo service postgresql start

# Verify PostgreSQL is running
pg_isready
```

### Step 2: Run Database Migrations

The application will automatically create the database if it doesn't exist. Run migrations to create tables:

```bash
go run cmd/api/main.go migrate
```

This creates:

- `orders` table: Stores order information
- `order_items` table: Stores order line items

### Step 3: Seed Sample Data (Optional)

Generate 50 sample orders for testing:

```bash
go run cmd/api/main.go seed
```

Sample data includes:

- 50 orders with random statuses (pending, processing, shipped, delivered, cancelled)
- 1-5 items per order
- Random amounts between $10-$1000
- Orders distributed over the past year

### Step 4: Start the API Server

```bash
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`

You should see:

```
Database connection established successfully
Server starting on localhost:8080
```

## API Endpoints

Base URL: `http://localhost:8080/api/v1`

### 1. Create Order

**Endpoint**: `POST /api/v1/orders`

**Description**: Create a new order with items

**Request Headers**:

```
Content-Type: application/json
```

**Request Body**:

```json
{
  "customer_id": "cust-123",
  "total_amount": 199.99,
  "status": "pending",
  "items": [
    {
      "product_id": "prod-001",
      "quantity": 2,
      "price": 49.99
    },
    {
      "product_id": "prod-002",
      "quantity": 1,
      "price": 99.99
    }
  ]
}
```

**Response** (201 Created):

```json
{
  "id": 1,
  "customer_id": "cust-123",
  "total_amount": 199.99,
  "status": "pending",
  "items": [
    {
      "id": 1,
      "order_id": 1,
      "product_id": "prod-001",
      "quantity": 2,
      "price": 49.99
    },
    {
      "id": 2,
      "order_id": 1,
      "product_id": "prod-002",
      "quantity": 1,
      "price": 99.99
    }
  ],
  "created_at": "2026-02-09T10:30:00Z",
  "updated_at": "2026-02-09T10:30:00Z"
}
```

**Validation Rules**:

- `customer_id`: Required, non-empty string
- `total_amount`: Required, must be > 0
- `status`: Required, valid values: `pending`, `processing`, `shipped`, `delivered`, `cancelled`
- `items`: Optional array of order items

**Error Response** (400 Bad Request):

```json
{
  "error": "customer_id is required"
}
```

### 2. List Orders

**Endpoint**: `GET /api/v1/orders`

**Description**: Retrieve a paginated list of orders with optional filters

**Query Parameters**:

| Parameter     | Type    | Description                            | Example                            |
| ------------- | ------- | -------------------------------------- | ---------------------------------- |
| `page`        | integer | Page number (default: 1)               | `?page=2`                          |
| `limit`       | integer | Items per page (default: 10, max: 100) | `?limit=20`                        |
| `status`      | string  | Filter by order status                 | `?status=pending`                  |
| `customer_id` | string  | Filter by customer ID                  | `?customer_id=cust-123`            |
| `amount`      | float   | Amount range (min,max)                 | `?amount=50,500`                   |
| `dateRange`   | date    | Date range (from,to, YYYY-MM-DD)       | `?dateRange=2026-01-01,2026-12-31` |

**Response** (200 OK):

```json
{
  "orders": [
    {
      "id": 1,
      "customer_id": "cust-123",
      "total_amount": 199.99,
      "status": "pending",
      "created_at": "2026-02-09T10:30:00Z",
      "updated_at": "2026-02-09T10:30:00Z"
    },
    {
      "id": 2,
      "customer_id": "cust-456",
      "total_amount": 299.99,
      "status": "shipped",
      "created_at": "2026-02-08T14:20:00Z",
      "updated_at": "2026-02-08T14:20:00Z"
    }
  ],
  "total": 150,
  "page": 1,
  "limit": 10,
  "offset": 0,
  "total_pages": 15
}
```

**Filter Combinations**:

- All filters can be combined
- Invalid parameter values return `400 Bad Request`
- Date format must be `YYYY-MM-DD`
- Amount filters accept decimal values
- Amount range must be `min <= max`
- Date range must be `from <= to`

## Usage Examples

### Example 1: Create a Simple Order

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "cust-001",
    "total_amount": 149.99,
    "status": "pending",
    "items": [
      {
        "product_id": "laptop-15",
        "quantity": 1,
        "price": 149.99
      }
    ]
  }'
```

### Example 2: List All Orders (Default Pagination)

```bash
curl http://localhost:8080/api/v1/orders
```

### Example 3: List Orders with Custom Pagination

```bash
# Get page 2 with 20 orders per page
curl "http://localhost:8080/api/v1/orders?page=2&limit=20"
```

**Response** (200 OK):

```json
{
  "orders": [],
  "total": 150,
  "page": 2,
  "limit": 20,
  "offset": 20,
  "total_pages": 8
}
```

### Example 4: Filter by Status

```bash
# Get all pending orders
curl "http://localhost:8080/api/v1/orders?status=pending"

# Get all shipped orders
curl "http://localhost:8080/api/v1/orders?status=shipped"
```

### Example 5: Filter by Customer ID

```bash
curl "http://localhost:8080/api/v1/orders?customer_id=cust-123"
```

### Example 6: Filter by Amount Range

```bash
# Orders between $50 and $200
curl "http://localhost:8080/api/v1/orders?amount=50,200"

# Orders exactly $1000
curl "http://localhost:8080/api/v1/orders?amount=1000"
```

### Example 7: Filter by Date Range

```bash
# Orders in January 2026
curl "http://localhost:8080/api/v1/orders?dateRange=2026-01-01,2026-01-31"

# Orders from last 7 days
curl "http://localhost:8080/api/v1/orders?dateRange=2026-02-02,2026-02-09"
```

### Example 8: Combined Filters

```bash
# Pending orders for customer cust-123 between $100-$500 in February 2026
curl "http://localhost:8080/api/v1/orders?status=pending&customer_id=cust-123&amount=100,500&dateRange=2026-02-01,2026-02-28&page=1&limit=10"
```

**Response** (200 OK):

```json
{
  "orders": [
    {
      "id": 10,
      "customer_id": "cust-123",
      "total_amount": 250.0,
      "status": "pending",
      "created_at": "2026-02-10T10:30:00Z",
      "updated_at": "2026-02-10T10:30:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10,
  "offset": 0,
  "total_pages": 1
}
```

### Example 9: Using with HTTPie (Alternative)

```bash
# Install HTTPie: pip install httpie

# Create order
http POST localhost:8080/api/v1/orders \
  customer_id=cust-001 \
  total_amount:=99.99 \
  status=pending

# List orders
http GET localhost:8080/api/v1/orders status==pending page==1 limit==10
```

### Example 10: Using with JavaScript/Fetch

```javascript
// Create order
const createOrder = async () => {
  const response = await fetch("http://localhost:8080/api/v1/orders", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      customer_id: "cust-001",
      total_amount: 299.99,
      status: "pending",
      items: [{ product_id: "prod-123", quantity: 2, price: 149.99 }],
    }),
  });
  const data = await response.json();
  console.log(data);
};

// List orders with filters
const listOrders = async () => {
  const params = new URLSearchParams({
    page: 1,
    limit: 20,
    status: "pending",
    min_amount: 100,
  });

  const response = await fetch(`http://localhost:8080/api/v1/orders?${params}`);
  const data = await response.json();
  console.log(data);
};
```

## Testing

The project includes comprehensive test coverage for the handlers package.

### Run All Tests

```bash
go test ./... -v
```

### Run Tests with Coverage

```bash
go test ./internal/handlers -cover
```

### Generate HTML Coverage Report

```bash
go test ./internal/handlers -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Statistics

- **Total Tests**: 19 test cases
- **Coverage**: 100% for handlers package
- **Test Categories**:
  - CRUD Operations (3 tests)
  - Pagination (4 tests)
  - Filtering (8 tests)
  - Middleware & Routes (4 tests)

### Test Cases Overview

**Order Creation Tests**:

- Valid order creation
- Invalid JSON body handling
- Service validation errors

**Pagination Tests**:

- Default pagination
- Custom pagination
- Maximum limit enforcement
- Negative/zero value handling

**Filter Tests**:

- Status filtering
- Customer ID filtering
- Amount range filtering
- Date range filtering
- Combined filters
- Invalid parameter handling

**Infrastructure Tests**:

- Route setup
- CORS middleware
- Logging middleware
