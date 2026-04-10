package postgres

import (
	"context"
	"database/sql"

	"payment-service/internal/domain"
	"payment-service/internal/usecase"
)

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) usecase.PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Save(ctx context.Context, p *domain.Payment) error {
	const q = `
		INSERT INTO payments (id, order_id, transaction_id, amount, status)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, q, p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status)
	return err
}

func (r *paymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	const q = `
		SELECT id, order_id, transaction_id, amount, status
		FROM payments WHERE order_id = $1`
	p := &domain.Payment{}
	err := r.db.QueryRowContext(ctx, q, orderID).
		Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPaymentNotFound
	}
	return p, err
}
