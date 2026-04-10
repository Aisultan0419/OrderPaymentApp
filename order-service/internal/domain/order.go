package domain

import (
	"errors"
	"time"
)

const (
	StatusPending   = "Pending"
	StatusPaid      = "Paid"
	StatusFailed    = "Failed"
	StatusCancelled = "Cancelled"
)

// Domain errors
var (
	ErrOrderNotFound = errors.New("order not found")
	ErrInvalidAmount = errors.New("amount must be greater than 0")
	ErrCannotCancel  = errors.New("only pending orders can be cancelled")
	ErrMissingFields = errors.New("customer_id and item_name are required")
)

type Order struct {
	ID             string
	CustomerID     string
	ItemName       string
	Amount         int64 // in cents, e.g. 15000 = $150.00
	Status         string
	IdempotencyKey string
	CreatedAt      time.Time
}

func NewOrder(id, customerID, itemName string, amount int64, idempotencyKey string) (*Order, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if customerID == "" || itemName == "" {
		return nil, ErrMissingFields
	}
	return &Order{
		ID:             id,
		CustomerID:     customerID,
		ItemName:       itemName,
		Amount:         amount,
		Status:         StatusPending,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now().UTC(),
	}, nil
}

func (o *Order) Cancel() error {
	if o.Status != StatusPending {
		return ErrCannotCancel
	}
	o.Status = StatusCancelled
	return nil
}
