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
		v1.POST("/payments", h.AuthorizePayment)
		v1.GET("/payments/:order_id", h.GetPaymentByOrderID)
	}

	return r
}
