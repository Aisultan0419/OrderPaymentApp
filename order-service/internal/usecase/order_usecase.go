package usecase

import (
	"context"

	"github.com/google/uuid"

	"order-service/internal/domain"
)

type OrderUseCase struct {
	repo          OrderRepository
	paymentClient PaymentClient
}

func NewOrderUseCase(repo OrderRepository, paymentClient PaymentClient) *OrderUseCase {
	return &OrderUseCase{
		repo:          repo,
		paymentClient: paymentClient,
	}
}

type CreateOrderInput struct {
	CustomerID     string
	ItemName       string
	Amount         int64
	IdempotencyKey string
	SkipPayment    bool
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, in CreateOrderInput) (*domain.Order, error) {
	if in.IdempotencyKey != "" {
		if existing, err := uc.repo.GetByIdempotencyKey(ctx, in.IdempotencyKey); err == nil {
			return existing, nil
		}
	}

	order, err := domain.NewOrder(uuid.New().String(), in.CustomerID, in.ItemName, in.Amount, in.IdempotencyKey)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	if in.SkipPayment {
		return order, nil
	}

	result, err := uc.paymentClient.AuthorizePayment(ctx, order.ID, order.Amount)
	if err != nil {
		_ = uc.repo.UpdateStatus(ctx, order.ID, domain.StatusFailed)
		order.Status = domain.StatusFailed
		return order, err
	}

	if result.Status == "Authorized" {
		order.Status = domain.StatusPaid
	} else {
		order.Status = domain.StatusFailed
	}

	_ = uc.repo.UpdateStatus(ctx, order.ID, order.Status)
	return order, nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := order.Cancel(); err != nil {
		return nil, err
	}

	if err := uc.repo.UpdateStatus(ctx, order.ID, order.Status); err != nil {
		return nil, err
	}

	return order, nil
}
