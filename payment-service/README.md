# Payment Service

Processes payment authorizations. Part of the Order & Payment microservices platform.

## Architecture

```
cmd/payment-service/main.go          ← Composition Root (manual DI only)
internal/
  domain/payment.go                  ← Pure domain entity, zero external deps
  usecase/
    interfaces.go                    ← Port: PaymentRepository
    payment_usecase.go               ← Business logic: Authorize / GetByOrderID
  repository/postgres/
    payment_repository.go            ← PostgreSQL adapter (implements PaymentRepository port)
  transport/http/
    handler.go                       ← Thin Gin handlers (parse → usecase → respond)
    router.go                        ← Route registration + Swagger mount
docs/docs.go                         ← Swagger spec
migrations/001_create_payments.sql   ← DB schema
```

### Dependency Flow

```
main.go
  └─ NewPaymentUseCase(repo)
       └─ repo → postgres.NewPaymentRepository(db)   [PaymentRepository port]
```

## Bounded Context

The Payment Service owns:
- Payment authorization and status
- Its own PostgreSQL database (`payments_db`) — never touches `orders_db`
- No knowledge of Order details beyond the `order_id` foreign key reference

## Business Rules

| Rule                          | Behaviour                                      |
|-------------------------------|------------------------------------------------|
| `amount <= 0`                 | Returns 400 Bad Request                        |
| `amount > 100000` (> $1000)   | Payment stored as **Declined**, no transaction ID |
| `amount <= 100000`            | Payment stored as **Authorized** with unique transaction ID |

All rules live in `domain.NewPayment()` — not in the handler.

## Endpoints

| Method | Path                        | Description                       |
|--------|-----------------------------|-----------------------------------|
| POST   | /payments                   | Authorize a payment               |
| GET    | /payments/:order_id         | Get payment status by order ID    |
| GET    | /swagger/index.html         | Swagger UI                        |

## Environment Variables

| Variable    | Default      | Description              |
|-------------|--------------|--------------------------|
| DB_HOST     | localhost    | PostgreSQL host          |
| DB_PORT     | 5432         | PostgreSQL port          |
| DB_USER     | postgres     | PostgreSQL user          |
| DB_PASSWORD | postgres     | PostgreSQL password      |
| DB_NAME     | payments_db  | PostgreSQL database name |
| PORT        | 8081         | HTTP listen port         |

## Running

```bash
# 1. Apply migration
psql -U postgres -d payments_db -f migrations/001_create_payments.sql

# 2. Install dependencies
go mod tidy

# 3. Run
go run ./cmd/payment-service
```
