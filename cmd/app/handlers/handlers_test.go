package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"educabot.com/bookshop/internal/books/providers/mockImpls"
	"educabot.com/bookshop/internal/books/services"
	appErr "educabot.com/bookshop/internal/platform/errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockFailService struct {
	Err error
}

func (m *MockFailService) GetMetrics(ctx context.Context, author string) (map[string]interface{}, error) {
	return nil, m.Err
}

func TestGetMetrics_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockProvider := mockImpls.NewMockBooksProvider()

	service := services.NewMetricsService(mockProvider)

	handler := NewMetricsHandler(service)

	r := gin.Default()
	r.GET("/", handler.GetMetrics())

	req := httptest.NewRequest(http.MethodGet, "/?author=Alan+Donovan", nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	var resBody map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &resBody)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, 11000, int(resBody["mean_units_sold"].(float64)))
	assert.Equal(t, "The Go Programming Language", resBody["cheapest_book"])
	assert.Equal(t, 1, int(resBody["books_written_by_author"].(float64)))
}

func TestGetMetrics_ErrorExternalService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock que fuerza un error del servicio externo
	mockService := &MockFailService{Err: appErr.ErrExternalService}

	handler := NewMetricsHandler(mockService)

	r := gin.Default()
	r.GET("/", handler.GetMetrics())

	req := httptest.NewRequest(http.MethodGet, "/?author=Tolkien", nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	var resBody map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &resBody)

	assert.Equal(t, http.StatusBadGateway, res.Code)
	assert.Equal(t, "external service error", resBody["message"])
}

func TestGetMetrics_MissingAuthor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockProvider := mockImpls.NewMockBooksProvider()
	service := services.NewMetricsService(mockProvider)
	handler := NewMetricsHandler(service)

	r := gin.Default()
	r.GET("/", handler.GetMetrics())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	var resBody map[string]interface{}
	json.Unmarshal(res.Body.Bytes(), &resBody)

	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, "missing required query parameter: author", resBody["message"])
}
