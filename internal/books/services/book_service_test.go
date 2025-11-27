package services

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"educabot.com/bookshop/internal/books/models"
	"educabot.com/bookshop/internal/books/providers/mockImpls"
	appErr "educabot.com/bookshop/internal/platform/errors"
	"github.com/stretchr/testify/assert"
)

type mockErrorBooksProvider struct {
	err error
}

func (m *mockErrorBooksProvider) GetBooks(ctx context.Context) ([]models.Book, error) {
	return nil, m.err
}

func TestMetricsService_GetMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		provider   models.BooksProvider
		author     string
		assertErr  func(t *testing.T, err error)
		assertData func(t *testing.T, data map[string]interface{})
	}{
		{
			name:     "success_returns_metrics",
			provider: mockImpls.NewMockBooksProvider(),
			author:   "Alan Donovan",
			assertErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			assertData: func(t *testing.T, data map[string]interface{}) {
				mean, ok := data["mean_units_sold"].(uint)
				assert.True(t, ok)
				assert.Equal(t, uint(11000), mean)

				cheapest, ok := data["cheapest_book"].(string)
				assert.True(t, ok)
				assert.Equal(t, "The Go Programming Language", cheapest)

				written, ok := data["books_written_by_author"].(uint)
				assert.True(t, ok)
				assert.Equal(t, uint(1), written)
			},
		},
		{
			name: "error_from_provider_propagated",
			provider: &mockErrorBooksProvider{
				err: appErr.ErrExternalService,
			},
			author: "any",
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, appErr.ErrExternalService))
			},
			assertData: func(t *testing.T, data map[string]interface{}) {
				assert.Nil(t, data)
			},
		},
		{
			name: "wrapped_error_from_provider_preserves_cause",
			provider: &mockErrorBooksProvider{
				err: fmt.Errorf("wrapped: %w", appErr.ErrExternalService),
			},
			author: "any",
			assertErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, appErr.ErrExternalService))
			},
			assertData: func(t *testing.T, data map[string]interface{}) {
				assert.Nil(t, data)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := NewMetricsService(tt.provider)

			ctx := context.Background()
			data, err := service.GetMetrics(ctx, tt.author)

			if tt.assertErr != nil {
				tt.assertErr(t, err)
			}

			if tt.assertData != nil {
				tt.assertData(t, data)
			}
		})
	}
}
