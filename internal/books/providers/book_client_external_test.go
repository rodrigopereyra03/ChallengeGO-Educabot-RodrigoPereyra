package providers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	appErr "educabot.com/bookshop/internal/platform/errors"
	"educabot.com/bookshop/internal/platform/resilience"
	"github.com/stretchr/testify/assert"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newTestClient(fn roundTripperFunc) *http.Client {
	return &http.Client{Transport: fn}
}

func TestHttpBooksProvider_GetBooks(t *testing.T) {
	t.Parallel()

	t.Run("success_maps_response_to_model", func(t *testing.T) {
		t.Parallel()

		client := newTestClient(func(r *http.Request) (*http.Response, error) {
			w := httptest.NewRecorder()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id":1,"name":"Book","author":"Author","units_sold":10,"price":20}]`))
			return w.Result(), nil
		})

		provider := NewExternalBooksProvider(client, "http://example.com")

		books, err := provider.GetBooks(context.Background())

		assert.NoError(t, err)
		assert.Len(t, books, 1)
		assert.Equal(t, uint(10), books[0].UnitsSold)
		assert.Equal(t, "Book", books[0].Name)
	})

	t.Run("network_error_returns_external_service_error", func(t *testing.T) {
		t.Parallel()

		client := newTestClient(func(r *http.Request) (*http.Response, error) {
			return nil, errors.New("dial error")
		})

		provider := NewExternalBooksProvider(client, "http://example.com")

		_, err := provider.GetBooks(context.Background())

		assert.Error(t, err)
		assert.True(t, errors.Is(err, appErr.ErrExternalService))
	})

	t.Run("json_decode_error_returns_invalid_response", func(t *testing.T) {
		t.Parallel()

		client := newTestClient(func(r *http.Request) (*http.Response, error) {
			w := httptest.NewRecorder()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{invalid-json`))
			return w.Result(), nil
		})

		provider := NewExternalBooksProvider(client, "http://example.com")

		_, err := provider.GetBooks(context.Background())

		assert.Error(t, err)
		assert.True(t, errors.Is(err, appErr.ErrInvalidResponse))
	})

	t.Run("http_422_returns_invalid_response", func(t *testing.T) {
		t.Parallel()

		client := newTestClient(func(r *http.Request) (*http.Response, error) {
			w := httptest.NewRecorder()
			w.WriteHeader(http.StatusUnprocessableEntity)
			return w.Result(), nil
		})

		provider := NewExternalBooksProvider(client, "http://example.com")

		_, err := provider.GetBooks(context.Background())

		assert.Error(t, err)
		assert.True(t, errors.Is(err, appErr.ErrExternalService) || errors.Is(err, appErr.ErrInvalidResponse))
	})

	t.Run("http_500_triggers_retries_and_returns_external_service_error", func(t *testing.T) {
		t.Parallel()

		var calls int

		client := newTestClient(func(r *http.Request) (*http.Response, error) {
			calls++
			w := httptest.NewRecorder()
			w.WriteHeader(http.StatusInternalServerError)
			return w.Result(), nil
		})

		provider := NewExternalBooksProvider(client, "http://example.com")

		_, err := provider.GetBooks(context.Background())

		assert.Error(t, err)
		assert.True(t, errors.Is(err, appErr.ErrExternalService))
		// Debe haber reintentos (3 intentos en total)
		assert.Equal(t, 3, calls)
	})

	t.Run("circuit_breaker_opens_and_fails_fast", func(t *testing.T) {
		t.Parallel()

		var calls int

		client := newTestClient(func(r *http.Request) (*http.Response, error) {
			calls++
			w := httptest.NewRecorder()
			w.WriteHeader(http.StatusInternalServerError)
			return w.Result(), nil
		})

		provider := NewExternalBooksProvider(client, "http://example.com")

		// 3 llamadas que fallan y deberían contribuir al contador del breaker
		for i := 0; i < 3; i++ {
			_, _ = provider.GetBooks(context.Background())
		}

		callsBefore := calls

		// Esta llamada debería ser rechazada por el circuit breaker (fail-fast)
		_, err := provider.GetBooks(context.Background())

		assert.Error(t, err)
		assert.True(t, errors.Is(err, appErr.ErrExternalService))
		assert.True(t, errors.Is(err, resilience.ErrCircuitOpen))

		// Como el breaker está abierto, no debería haberse incrementado el contador de llamadas HTTP
		assert.Equal(t, callsBefore, calls)
	})
}
