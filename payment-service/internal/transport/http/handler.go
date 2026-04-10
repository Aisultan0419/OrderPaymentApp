package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"payment-service/internal/domain"
	"payment-service/internal/usecase"
)

// Handler is the thin HTTP delivery layer for the Payment Service.
type Handler struct {
	uc *usecase.PaymentUseCase
}

func NewHandler(uc *usecase.PaymentUseCase) *Handler {
	return &Handler{uc: uc}
}

type authorizePaymentRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	Amount  int64  `json:"amount"   binding:"required,gt=0"`
}

type paymentResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func toPaymentResponse(p *domain.Payment) paymentResponse {
	return paymentResponse{
		ID:            p.ID,
		OrderID:       p.OrderID,
		TransactionID: p.TransactionID,
		Amount:        p.Amount,
		Status:        p.Status,
	}
}

// AuthorizePayment godoc
// @Summary      Authorize a payment
// @Description  Processes a payment for an order. Amounts > 100000 cents ($1000) are automatically Declined.
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        request  body      authorizePaymentRequest  true  "Payment payload"
// @Success      201      {object}  paymentResponse
// @Failure      400      {object}  errorResponse
// @Router       /payments [post]
func (h *Handler) AuthorizePayment(c *gin.Context) {
	var req authorizePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	payment, err := h.uc.Authorize(c.Request.Context(), usecase.AuthorizeInput{
		OrderID: req.OrderID,
		Amount:  req.Amount,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toPaymentResponse(payment))
}

// GetPaymentByOrderID godoc
// @Summary      Get payment by order ID
// @Description  Returns the payment record associated with the given order ID.
// @Tags         payments
// @Produce      json
// @Param        order_id  path      string  true  "Order ID"
// @Success      200       {object}  paymentResponse
// @Failure      404       {object}  errorResponse
// @Router       /payments/{order_id} [get]
func (h *Handler) GetPaymentByOrderID(c *gin.Context) {
	payment, err := h.uc.GetByOrderID(c.Request.Context(), c.Param("order_id"))
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, errorResponse{Error: "payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toPaymentResponse(payment))
}
