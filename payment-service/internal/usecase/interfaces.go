package usecase

import (
	"context"

	"payment-service/internal/domain"
)

type PaymentRepository interface {
	Save(ctx context.Context, p *domain.Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
}
