package usecase

import (
	"context"
	"fmt"

	"order-service/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id, status string) error
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error)
}

type PaymentResult struct {
	TransactionID string
	Status        string // "Authorized" or "Declined"
}

type PaymentClient interface {
	AuthorizePayment(ctx context.Context, orderID string, amount int64) (*PaymentResult, error)
}

type ErrServiceUnavailable struct {
	Service string
}

func (e *ErrServiceUnavailable) Error() string {
	return fmt.Sprintf("%s service is unavailable", e.Service)
}
