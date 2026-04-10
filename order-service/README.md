# Order Service

Manages customer orders. Part of the Order & Payment microservices platform.

## Architecture

```
cmd/order-service/main.go          ← Composition Root (manual DI only)
internal/
  domain/order.go                  ← Pure domain entity, zero external deps
  usecase/
    interfaces.go                  ← Ports: OrderRepository, PaymentClient
    order_usecase.go               ← Business logic: Create / Get / Cancel
  repository/postgres/
    order_repository.go            ← PostgreSQL adapter (implements OrderRepository port)
  client/
    payment_client.go              ← HTTP adapter (implements PaymentClient port, 2s timeout)
  transport/http/
    handler.go                     ← Thin Gin handlers (parse → usecase → respond)
    router.go                      ← Route registration + Swagger mount
docs/docs.go                       ← Swagger spec
migrations/001_create_orders.sql   ← DB schema
```

### Dependency Flow

```
main.go
  └─ NewOrderUseCase(repo, paymentClient)
       ├─ repo           → postgres.NewOrderRepository(db)   [OrderRepository port]
       └─ paymentClient  → client.NewPaymentServiceClient()  [PaymentClient port]
```

Nothing in `domain/` or `usecase/` imports Gin, `database/sql`, or `net/http`.
All concrete dependencies flow inward from `main.go` only.

## Bounded Context

The Order Service owns:
- Order lifecycle: Pending → Paid / Failed / Cancelled
- Its own PostgreSQL database (`orders_db`) — never touches `payments_db`
- Communication with Payment Service **only** through the `PaymentClient` interface

## Failure Handling

If the Payment Service is unreachable or returns a 5xx:

1. The HTTP client times out after **2 seconds** (hard limit set on `http.Client`).
2. The Order is marked **"Failed"** in the database (consistent state, never "limbo").
3. The caller receives **HTTP 503 Service Unavailable**.

Design decision: marking as "Failed" (rather than leaving as "Pending") means the
order state is always deterministic after a create attempt. A Pending order always
means "waiting for payment decision", not "payment call crashed".

## Idempotency (Bonus)

Send `Idempotency-Key: <your-unique-key>` header with `POST /orders`.
If an order with that key already exists, it is returned immediately without
creating a duplicate order or calling the Payment Service again.

## Endpoints

| Method | Path                  | Description                        |
|--------|-----------------------|------------------------------------|
| POST   | /orders               | Create order + authorize payment   |
| GET    | /orders/:id           | Get order by ID                    |
| PATCH  | /orders/:id/cancel    | Cancel a Pending order             |
| GET    | /swagger/index.html   | Swagger UI                         |

## Environment Variables

| Variable              | Default                   | Description                      |
|-----------------------|---------------------------|----------------------------------|
| DB_HOST               | localhost                 | PostgreSQL host                  |
| DB_PORT               | 5432                      | PostgreSQL port                  |
| DB_USER               | postgres                  | PostgreSQL user                  |
| DB_PASSWORD           | postgres                  | PostgreSQL password              |
| DB_NAME               | orders_db                 | PostgreSQL database name         |
| PORT                  | 8080                      | HTTP listen port                 |
| PAYMENT_SERVICE_URL   | http://localhost:8081     | Base URL of the Payment Service  |

## Running

```bash
# 1. Apply migration
psql -U postgres -d orders_db -f migrations/001_create_orders.sql

# 2. Install dependencies
go mod tidy

# 3. Run
go run ./cmd/order-service
```
