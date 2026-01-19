package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"subscription_service/internal/http/middleware"
)

func NewRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	v1 := r.Group("/api/v1")
	{
		v1.POST("/subscriptions", h.Create)
		v1.GET("/subscriptions/:id", h.GetByID)
		v1.PATCH("/subscriptions/:id", h.Update)
		v1.DELETE("/subscriptions/:id", h.Delete)
		v1.GET("/subscriptions", h.List)

		v1.GET("/subscriptions/total", h.Total)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
