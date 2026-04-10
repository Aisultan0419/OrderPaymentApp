package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(h *Handler) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/")
	{
		v1.POST("/orders/pending", h.CreatePendingOrder)
		v1.POST("/orders", h.CreateOrder)
		v1.GET("/orders/:id", h.GetOrder)
		v1.PATCH("/orders/:id/cancel", h.CancelOrder)
		v1.PATCH("/orders/:id/status", h.UpdateOrderStatus)
	}

	return r
}
