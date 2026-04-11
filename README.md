# 🛒 Order & Payment Platform — AP2 Assignment 2

> Microservices platform built with **Clean Architecture** in Go.  
> Assignment 2 migrates inter-service communication from REST to **gRPC**, adopting a Contract-First approach with automated code generation via GitHub Actions.

---

## 🗂 Repositories

| Repository | Purpose |
|---|---|
| [`ap2-protos`](https://github.com/Aisultan0419/ap2-protos) | Source of truth — `.proto` contract definitions |
| [`ap2-gen`](https://github.com/Aisultan0419/ap2-gen) | Auto-generated `.pb.go` files (triggered by GitHub Actions) |
| This repo | `order-service` + `payment-service` implementations |

---


## 🧱 Clean Architecture Layers

```
Domain       → pure structs + business invariants, zero external deps
    ↑
UseCase      → orchestration logic, depends on interfaces (ports) only
    ↑
Repository   → PostgreSQL adapter (implements repo port)
Client       → gRPC adapter (implements PaymentClient port)   [Order Service only]
    ↑
Transport    → thin Gin handlers (HTTP) + gRPC server handlers
    ↑
main.go      → Composition Root: wires all layers together with manual DI
```

> **Nothing in `domain/` or `usecase/` imports Gin, `database/sql`, `net/http`, or gRPC.**  
> Business logic from Assignment 1 is untouched — only the transport layer changed.

---

## ⚡ Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL on `localhost:5432`

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

### 4 — Run (two terminals)

```bash
# Terminal 1 — start Payment Service first
cd payment-service
go run ./cmd/payment-service

# Terminal 2 — start Order Service
cd order-service
go run ./cmd/order-service
```

### 5 — Swagger UI

| Service | URL |
|---|---|
| Order Service | http://localhost:8080/swagger/index.html |
| Payment Service | http://localhost:8081/swagger/index.html |

---

## 🧪 Testing Scenarios

### Create a paid order (triggers gRPC call)

```bash
curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust-1","item_name":"Laptop","amount":15000}' | jq
# → status: "Paid"
# → payment-service terminal shows: [gRPC] method=/payment.PaymentService/ProcessPayment duration=...
```

### Declined payment (amount > $1000)

```bash
curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust-1","item_name":"Car","amount":200000}' | jq
# → status: "Failed"
```

### Real-time order streaming

```bash
# Terminal A — subscribe to order updates
go run ./order-service/cmd/stream-client/main.go <ORDER_ID>

# Terminal B — update order status (triggers stream push)
curl -s -X PATCH http://localhost:8080/orders/<ORDER_ID>/status \
  -H "Content-Type: application/json" \
  -d '{"status":"Paid"}'

# Terminal A output:
# Watching order <ORDER_ID>...
# Status update → Pending
# Status update → Paid
# Stream closed.
```

### Cancel a pending order

```bash
# Create pending order (skips payment)
curl -s -X POST http://localhost:8080/orders/pending \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust-1","item_name":"Book","amount":500}' | jq

# Cancel it
curl -s -X PATCH http://localhost:8080/orders/<ORDER_ID>/cancel | jq
# → status: "Cancelled"
```

---

## 🔑 Key Design Decisions

| Topic | Decision | Rationale |
|---|---|---|
| Inter-service protocol | gRPC | Strong typing, generated interfaces, no manual JSON marshaling |
| Contract management | Remote generation via GitHub Actions | Contract-First: proto is the source of truth |
| Dependency versioning | Tagged releases (`v1.0.0`) on `ap2-gen` | Reproducible builds, easy rollback |
| Streaming implementation | DB polling (1s interval) | Real data, no fake `time.Sleep` loops |
| Logging | gRPC Unary Interceptor | Non-invasive, logs method + duration for every RPC |
| Clean Architecture | UseCase and Domain unchanged | Only transport layer modified — business logic preserved |
| Money type | `int64` (cents) | Avoids floating-point rounding errors |
| Hardcoded addresses | None — all via `.env` | Required by assignment; no ports/IPs in source code |