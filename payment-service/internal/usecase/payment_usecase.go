package usecase

import (
	"context"

	"github.com/google/uuid"

	"payment-service/internal/domain"
)

type PaymentUseCase struct {
	repo PaymentRepository
}

func NewPaymentUseCase(repo PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

type AuthorizeInput struct {
	OrderID string
	Amount  int64
}

func (uc *PaymentUseCase) Authorize(ctx context.Context, in AuthorizeInput) (*domain.Payment, error) {
	payment, err := domain.NewPayment(
		uuid.New().String(),
		in.OrderID,
		uuid.New().String(), // transaction ID
		in.Amount,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (uc *PaymentUseCase) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	return uc.repo.GetByOrderID(ctx, orderID)
}
