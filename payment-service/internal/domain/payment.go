package domain

import "errors"

const (
	StatusAuthorized = "Authorized"
	StatusDeclined   = "Declined"

	MaxAuthorizedAmount int64 = 100000
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrInvalidAmount   = errors.New("amount must be greater than 0")
	ErrMissingOrderID  = errors.New("order_id is required")
)

type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64 // in cents
	Status        string
}

func NewPayment(id, orderID, transactionID string, amount int64) (*Payment, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if orderID == "" {
		return nil, ErrMissingOrderID
	}

	status := StatusAuthorized
	if amount > MaxAuthorizedAmount {
		status = StatusDeclined
		transactionID = "" // no transaction for declined payments
	}

	return &Payment{
		ID:            id,
		OrderID:       orderID,
		TransactionID: transactionID,
		Amount:        amount,
		Status:        status,
	}, nil
}
