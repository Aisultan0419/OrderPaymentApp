package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"order-service/internal/domain"
	"order-service/internal/usecase"
)

type Handler struct {
	uc *usecase.OrderUseCase
}

func NewHandler(uc *usecase.OrderUseCase) *Handler {
	return &Handler{uc: uc}
}

type createOrderRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	ItemName   string `json:"item_name"   binding:"required"`
	Amount     int64  `json:"amount"      binding:"required,gt=0"`
}

type orderResponse struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	ItemName   string `json:"item_name"`
	Amount     int64  `json:"amount"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func toOrderResponse(o *domain.Order) orderResponse {
	return orderResponse{
		ID:         o.ID,
		CustomerID: o.CustomerID,
		ItemName:   o.ItemName,
		Amount:     o.Amount,
		Status:     o.Status,
		CreatedAt:  o.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// CreateOrder godoc
// @Summary      Create a new order
// @Description  Creates an order, calls Payment Service to authorize, and returns the final status.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key  header    string             false  "Unique key to prevent duplicate orders"
// @Param        request          body      createOrderRequest  true   "Order payload"
// @Success      201              {object}  orderResponse
// @Failure      400              {object}  errorResponse
// @Failure      503              {object}  errorResponse
// @Router       /orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	input := usecase.CreateOrderInput{
		CustomerID:     req.CustomerID,
		ItemName:       req.ItemName,
		Amount:         req.Amount,
		IdempotencyKey: c.GetHeader("Idempotency-Key"),
	}

	order, err := h.uc.CreateOrder(c.Request.Context(), input)
	if err != nil {
		var svcErr *usecase.ErrServiceUnavailable
		if errors.As(err, &svcErr) {
			c.JSON(http.StatusServiceUnavailable, errorResponse{Error: "payment service unavailable, order marked as failed"})
			return
		}
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toOrderResponse(order))
}

// GetOrder godoc
// @Summary      Get order by ID
// @Description  Returns order details from the database.
// @Tags         orders
// @Produce      json
// @Param        id   path      string  true  "Order ID"
// @Success      200  {object}  orderResponse
// @Failure      404  {object}  errorResponse
// @Router       /orders/{id} [get]
func (h *Handler) GetOrder(c *gin.Context) {
	order, err := h.uc.GetOrder(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, errorResponse{Error: "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toOrderResponse(order))
}

// CancelOrder godoc
// @Summary      Cancel an order
// @Description  Cancels a Pending order. Paid orders cannot be cancelled.
// @Tags         orders
// @Produce      json
// @Param        id   path      string  true  "Order ID"
// @Success      200  {object}  orderResponse
// @Failure      400  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Router       /orders/{id}/cancel [patch]
func (h *Handler) CancelOrder(c *gin.Context) {
	order, err := h.uc.CancelOrder(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, errorResponse{Error: "order not found"})
			return
		}
		if errors.Is(err, domain.ErrCannotCancel) {
			c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toOrderResponse(order))
}

// CreatePendingOrder godoc
// @Summary      Create order that stays Pending (for cancel testing)
// @Description  Creates an order but intentionally skips payment — always returns Pending status.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        request  body      createOrderRequest  true  "Order payload"
// @Success      201      {object}  orderResponse
// @Failure      400      {object}  errorResponse
// @Router       /orders/pending [post]
func (h *Handler) CreatePendingOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	input := usecase.CreateOrderInput{
		CustomerID:  req.CustomerID,
		ItemName:    req.ItemName,
		Amount:      req.Amount,
		SkipPayment: true,
	}

	order, err := h.uc.CreateOrder(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toOrderResponse(order))
}
