# Order & Payment Platform — AP2 Assignment 1

Two-service platform built with **Clean Architecture** in Go.

## Architecture Diagram

```
+

│                         CLIENT (curl / Swagger)                      │

                            │ HTTP
                            ▼

│              ORDER SERVICE  :8080         │
│                                           │
│  transport/http  (Gin handlers)           │
│        │                                  │
│        ▼                                  │
│  usecase  (business logic)                │
│        │                   │              │
│        ▼                   ▼              │
│  repository           PaymentClient       │
│  (OrderRepository     (interface)         │
│   interface)               │              │
│        │                   │              │
│        ▼                   │              │
│  ┌───────────┐             │ HTTP POST    │
│  │ orders_db │             │ /payments    │
│  │ PostgreSQL│             │ (2s timeout) │
│  └───────────┘             │              │

                             │
                             ▼

│            PAYMENT SERVICE  :8081         │
│                                           │
│  transport/http  (Gin handlers)           │
│        │                                  │
│        ▼                                  │
│  usecase  (business logic)                │
│   • amount > 100000 → Declined            │
│        │                                  │
│        ▼                                  │
│  repository (PaymentRepository interface) │
│        │                                  │
│        ▼                                  │
│                            │
│  │ payments_db │                          │
│  │  PostgreSQL │                          │
│                           │

```

### Clean Architecture layers (per service)

```
Domain       → pure structs + business invariants, no external imports
  ↑
UseCase      → orchestration logic, depends on interfaces (ports) only
  ↑
Repository   → PostgreSQL adapter (implements repo port)
Client       → HTTP adapter (implements PaymentClient port)   [Order Service only]
  ↑
Transport    → thin Gin handlers: parse → usecase → respond
  ↑
main.go      → Composition Root: wires all layers together with manual DI
```

## Quick Start

### Prerequisites
- Go 1.22+
- PostgreSQL running on localhost:5432

### 1 — Create databases
```sql
psql -U postgres -c "CREATE DATABASE orders_db;"
psql -U postgres -c "CREATE DATABASE payments_db;"
```

### 2 — Run migrations
```bash
psql -U postgres -d orders_db   -f order-service/migrations/001_create_orders.sql
psql -U postgres -d payments_db -f payment-service/migrations/001_create_payments.sql
```

### 3 — Install dependencies
```bash
cd order-service   && go mod tidy && cd ..
cd payment-service && go mod tidy && cd ..
```

### 4 — (Optional) Regenerate Swagger docs
```bash
go install github.com/swaggo/swag/cmd/swag@latest

cd order-service   && swag init -g cmd/order-service/main.go   && cd ..
cd payment-service && swag init -g cmd/payment-service/main.go && cd ..
```

### 5 — Run (two terminals)
```bash
# Terminal 1 — start Payment Service first
cd payment-service
DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres \
  DB_NAME=payments_db PORT=8081 \
  go run ./cmd/payment-service

# Terminal 2 — start Order Service
cd order-service
DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres \
  DB_NAME=orders_db PORT=8080 \
  PAYMENT_SERVICE_URL=http://localhost:8081 \
  go run ./cmd/order-service
```

### 6 — Open Swagger UI
- Order Service:   http://localhost:8080/swagger/index.html
- Payment Service: http://localhost:8081/swagger/index.html

---

## Testing Scenarios

### Happy path — create a paid order
```bash
curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust-1","item_name":"Laptop","amount":15000}' | jq
# → status: "Paid"
```

### Declined payment (amount > 100000)
```bash
curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust-1","item_name":"Car","amount":200000}' | jq
# → status: "Failed"  (payment was Declined by Payment Service)
```

### Cancel a pending order
```bash
# 1. Stop the payment service so the order stays Pending (or use amount > limit)
curl -s -X PATCH http://localhost:8080/orders/<ORDER_ID>/cancel | jq
# → status: "Cancelled"

# 2. Try cancelling a Paid order
curl -s -X PATCH http://localhost:8080/orders/<PAID_ORDER_ID>/cancel | jq
# → 400: "only pending orders can be cancelled"
```

### Idempotency (bonus)
```bash
# Send twice with the same key — second call returns the same order, no duplicate
curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: my-unique-key-001" \
  -d '{"customer_id":"cust-1","item_name":"Book","amount":2000}' | jq
```

### Payment Service unavailable (503)
```bash
# Stop the payment service, then:
curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust-1","item_name":"Tablet","amount":5000}' | jq
# → HTTP 503, order marked "Failed" in DB, response within ~2 seconds
```

---

## Key Design Decisions

| Topic | Decision | Rationale |
|---|---|---|
| Money type | `int64` (cents) | Avoids all floating-point rounding errors |
| Payment timeout | 2s on `http.Client` | Order Service never hangs; assignment requirement |
| Unavailable → order status | `Failed` | Deterministic state; Pending always means "awaiting decision" |
| Cancelled Paid order | Returns 400 | Domain invariant enforced in `Order.Cancel()` |
| Declined payment | amount > 100000 | Business rule lives in `domain.NewPayment()`, not handler |
| Shared code | None | Each service has its own models; no `common/` package |
| DI approach | Manual in `main.go` | No DI framework; clean dependency graph, easy to trace |
