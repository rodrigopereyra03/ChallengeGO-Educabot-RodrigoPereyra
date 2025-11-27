package handlers

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"educabot.com/bookshop/internal/books/services"
	"github.com/gin-gonic/gin"
)

type GetMetricsRequest struct {
	Author string `form:"author"`
}

type MetricsHandler struct {
	service services.MetricsServiceI
}

func NewMetricsHandler(service services.MetricsServiceI) *MetricsHandler {
	return &MetricsHandler{service}
}

func (h *MetricsHandler) GetMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		author, _ := url.QueryUnescape(c.Query("author"))

		if strings.TrimSpace(author) == "" {
			writeJSON(c.Writer, http.StatusBadRequest, "missing required query parameter: author")
			return
		}

		result, err := h.service.GetMetrics(ctx, author)
		if err != nil {
			writeError(c.Writer, err)
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
