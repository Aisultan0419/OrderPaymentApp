# Payment Service

Processes payment authorizations. Part of the Order & Payment microservices platform.

> In Assignment 2, the Payment Service was extended with a **gRPC Server** interface.  
> It now accepts payment requests via both REST (for Swagger/testing) and **gRPC** (from Order Service).  
> A **Logging Interceptor** automatically logs every incoming gRPC call with method name and duration.

---

## Architecture

```
cmd/payment-service/main.go            ← Composition Root (manual DI only)
internal/
  domain/payment.go                    ← Pure domain entity, zero external deps
  usecase/
    interfaces.go                      ← Port: PaymentRepository
    payment_usecase.go                 ← Business logic: Authorize / GetByOrderID
  repository/postgres/
    payment_repository.go              ← PostgreSQL adapter (implements PaymentRepository port)
  transport/
    http/
      handler.go                       ← Thin Gin handlers (parse → usecase → respond)
      router.go                        ← Route registration + Swagger mount
    grpc/
      server.go                        ← gRPC server (ProcessPayment RPC)
      interceptor.go                   ← Logging interceptor (method name + duration)
docs/docs.go                           ← Swagger spec
migrations/001_create_payments.sql     ← DB schema
```

### Dependency Flow

```
main.go
  ├─ NewPaymentUseCase(repo)
  │    └─ repo → postgres.NewPaymentRepository(db)
  ├─ HTTP server  :8081   (Gin)
  └─ gRPC server  :9091   (with LoggingInterceptor)
```

Nothing in `domain/` or `usecase/` imports Gin, `database/sql`, or gRPC.

---

## Bounded Context

The Payment Service owns:
- Payment authorization and status
- Its own PostgreSQL database (`payments_db`) — never touches `orders_db`
- No knowledge of Order details beyond the `order_id` foreign key reference

---

## gRPC Logging Interceptor

Every incoming gRPC call is automatically logged to stdout:

```
[gRPC] method=/payment.PaymentService/ProcessPayment duration=2.341ms err=<nil>
[gRPC] method=/payment.PaymentService/ProcessPayment duration=1.102ms err=rpc error: code = InvalidArgument ...
```

The interceptor is a `grpc.UnaryServerInfo` middleware — it wraps every RPC handler without touching the UseCase or Domain layers.

---

## Business Rules

| Rule | Behaviour |
|---|---|
| `amount <= 0` | Returns `codes.InvalidArgument` (gRPC) / 400 (HTTP) |
| `order_id == ""` | Returns `codes.InvalidArgument` (gRPC) / 400 (HTTP) |
| `amount > 100000` (> $1000) | Payment stored as **Declined**, no transaction ID |
| `amount <= 100000` | Payment stored as **Authorized** with unique transaction ID |

All rules live in `domain.NewPayment()` — not in the handler or gRPC server.

---

## HTTP Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/payments` | Authorize a payment |
| `GET` | `/payments/:order_id` | Get payment status by order ID |
| `GET` | `/swagger/index.html` | Swagger UI |

## gRPC Endpoints

| RPC | Type | Description |
|---|---|---|
| `ProcessPayment` | Unary | Authorize a payment, returns status + transaction ID |

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `payments_db` | PostgreSQL database name |
| `PORT` | `8081` | HTTP listen port |
| `GRPC_PORT` | `9091` | gRPC server port |

---

## Running

```bash
# 1. Apply migration
psql -U postgres -d payments_db -f migrations/001_create_payments.sql

# 2. Install dependencies
go mod tidy

# 3. Run
go run ./cmd/payment-service
```

---

## Regenerate Swagger docs

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/payment-service/main.go
```