package postgres

import (
	"context"
	"database/sql"

	"order-service/internal/domain"
	"order-service/internal/usecase"
)

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) usecase.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, o *domain.Order) error {
	const q = `
		INSERT INTO orders (id, customer_id, item_name, amount, status, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	var idempotencyKey *string
	if o.IdempotencyKey != "" {
		idempotencyKey = &o.IdempotencyKey
	}

	_, err := r.db.ExecContext(ctx, q,
		o.ID, o.CustomerID, o.ItemName, o.Amount, o.Status, idempotencyKey, o.CreatedAt)
	return err
}

func (r *orderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	const q = `
		SELECT id, customer_id, item_name, amount, status, idempotency_key, created_at
		FROM orders WHERE id = $1`
	return r.scanOrder(r.db.QueryRowContext(ctx, q, id))
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id, status string) error {
	const q = `UPDATE orders SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, q, status, id)
	return err
}

func (r *orderRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error) {
	const q = `
		SELECT id, customer_id, item_name, amount, status, idempotency_key, created_at
		FROM orders WHERE idempotency_key = $1`
	return r.scanOrder(r.db.QueryRowContext(ctx, q, key))
}

func (r *orderRepository) scanOrder(row *sql.Row) (*domain.Order, error) {
	o := &domain.Order{}
	var idempotencyKey *string
	err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &idempotencyKey, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if idempotencyKey != nil {
		o.IdempotencyKey = *idempotencyKey
	}
	return o, err
}
