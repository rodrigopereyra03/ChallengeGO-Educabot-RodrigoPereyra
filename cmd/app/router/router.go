package router

import (
	"educabot.com/bookshop/cmd/app/handlers"
	"github.com/gin-gonic/gin"
)

type Router struct {
	bookHandler *handlers.MetricsHandler
}

func NewRouter(bookHandler *handlers.MetricsHandler) *Router {
	return &Router{bookHandler: bookHandler}
}

func (r *Router) Setup() *gin.Engine {
	engine := gin.Default()

	api := engine.Group("/api/v1")
	{
		api.GET("/books/metrics", r.bookHandler.GetMetrics())
	}

	return engine
}
