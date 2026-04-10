package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"order-service/internal/usecase"
)

type paymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type paymentResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

type PaymentServiceClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewPaymentServiceClient(baseURL string) usecase.PaymentClient {
	return &PaymentServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func (c *PaymentServiceClient) AuthorizePayment(ctx context.Context, orderID string, amount int64) (*usecase.PaymentResult, error) {
	body, err := json.Marshal(paymentRequest{OrderID: orderID, Amount: amount})
	if err != nil {
		return nil, fmt.Errorf("marshal payment request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/payments", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &usecase.ErrServiceUnavailable{Service: "payment"}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, &usecase.ErrServiceUnavailable{Service: "payment"}
	}

	var pr paymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("decode payment response: %w", err)
	}

	return &usecase.PaymentResult{
		TransactionID: pr.TransactionID,
		Status:        pr.Status,
	}, nil
}
