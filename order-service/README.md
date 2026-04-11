# Order Service

Manages customer orders. Part of the Order & Payment microservices platform.

> In Assignment 2, the internal call to Payment Service was migrated from REST to **gRPC**.  
> The Order Service also acts as a **gRPC Server** for real-time order status streaming.

---

## Architecture

```
cmd/order-service/main.go              ← Composition Root (manual DI only)
cmd/stream-client/main.go              ← CLI client to demo order status streaming
internal/
  domain/order.go                      ← Pure domain entity, zero external deps
  usecase/
    interfaces.go                      ← Ports: OrderRepository, PaymentClient
    order_usecase.go                   ← Business logic: Create / Get / Cancel / UpdateStatus
  repository/postgres/
    order_repository.go                ← PostgreSQL adapter (implements OrderRepository port)
  client/
    payment_client.go                  ← gRPC adapter (implements PaymentClient port)
  transport/
    http/
      handler.go                       ← Thin Gin handlers (parse → usecase → respond)
      router.go                        ← Route registration + Swagger mount
    grpc/
      server.go                        ← gRPC streaming server (SubscribeToOrderUpdates)
docs/docs.go                           ← Swagger spec
migrations/001_create_orders.sql       ← DB schema
```

### Dependency Flow

```
main.go
  ├─ NewPaymentGRPCClient(addr)         [PaymentClient port — gRPC]
  ├─ NewOrderUseCase(repo, paymentClient)
  │    ├─ repo           → postgres.NewOrderRepository(db)
  │    └─ paymentClient  → client.NewPaymentGRPCClient()
  ├─ HTTP server  :8080   (Gin)
  └─ gRPC server  :9090   (streaming)
```

Nothing in `domain/` or `usecase/` imports Gin, `database/sql`, or gRPC.

---

## Bounded Context

The Order Service owns:
- Order lifecycle: `Pending` → `Paid` / `Failed` / `Cancelled`
- Its own PostgreSQL database (`orders_db`) — never touches `payments_db`
- Communication with Payment Service **only** through the `PaymentClient` interface (gRPC)
- Real-time order status streaming to subscribed clients

---

## gRPC Streaming

The Order Service implements `SubscribeToOrderUpdates` — a Server-side Streaming RPC.

**How it works:**
1. Client connects and sends an `OrderRequest` with an `order_id`
2. Server polls the database every **1 second** for status changes
3. Whenever the status changes, it immediately pushes an `OrderStatusUpdate` to the client
4. Stream closes automatically when order reaches a terminal state (`Paid`, `Failed`, `Cancelled`)

This is tied to **real database changes** — updating the order status via `PATCH /orders/:id/status` triggers an immediate push to all active subscribers.

---

## Failure Handling

If the Payment Service is unreachable or returns a gRPC error:

1. The gRPC call returns `codes.Unavailable`
2. The Order is marked **"Failed"** in the database
3. The caller receives **HTTP 503 Service Unavailable**

---

## HTTP Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/orders` | Create order + authorize payment via gRPC |
| `POST` | `/orders/pending` | Create order, skip payment (for cancel testing) |
| `GET` | `/orders/:id` | Get order by ID |
| `PATCH` | `/orders/:id/cancel` | Cancel a Pending order |
| `PATCH` | `/orders/:id/status` | Update order status (triggers stream push) |
| `GET` | `/swagger/index.html` | Swagger UI |

## gRPC Endpoints

| RPC | Type | Description |
|---|---|---|
| `SubscribeToOrderUpdates` | Server-side Streaming | Stream real-time status updates for an order |

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | `postgres` | PostgreSQL password |
| `DB_NAME` | `orders_db` | PostgreSQL database name |
| `PORT` | `8080` | HTTP listen port |
| `GRPC_PORT` | `9090` | gRPC streaming server port |
| `PAYMENT_GRPC_ADDR` | `localhost:9091` | gRPC address of Payment Service |

---

## Running

```bash
# 1. Apply migration
psql -U postgres -d orders_db -f migrations/001_create_orders.sql

# 2. Install dependencies
go mod tidy

# 3. Run
go run ./cmd/order-service

# 4. (Optional) Run stream client demo
go run ./cmd/stream-client/main.go <ORDER_ID>
```

---

## Regenerate Swagger docs

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/order-service/main.go
```