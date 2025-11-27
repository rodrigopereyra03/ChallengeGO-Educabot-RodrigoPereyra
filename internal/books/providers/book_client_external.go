package providers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"educabot.com/bookshop/internal/books/models"
	appErr "educabot.com/bookshop/internal/platform/errors"
	"educabot.com/bookshop/internal/platform/resilience"
)

type HttpBooksProvider struct {
	client  *http.Client
	baseURL string
	cb      *resilience.CircuitBreaker
}

func NewExternalBooksProvider(client *http.Client, url string) *HttpBooksProvider {
	return &HttpBooksProvider{
		client:  client,
		baseURL: url,
		cb:      resilience.NewCircuitBreaker(3, 5*time.Second),
	}
}

func (p *HttpBooksProvider) GetBooks(ctx context.Context) ([]models.Book, error) {

	if err := p.cb.AllowRequest(); err != nil {
		return nil, errors.Join(appErr.ErrExternalService, err)
	}

	var lastErr error

	for attempt := 0; attempt < 3; attempt++ {

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL, nil)
		if err != nil {
			return nil, errors.Join(appErr.ErrExternalService, err)
		}

		resp, err := p.client.Do(req)
		if err != nil {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				lastErr = appErr.ErrTimeout
				break
			}

			lastErr = err
			time.Sleep(time.Duration(150*(1<<attempt)) * time.Millisecond) // backoff
			continue
		}

		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			_ = resp.Body.Close()
			lastErr = appErr.ErrExternalService
			time.Sleep(time.Duration(150*(1<<attempt)) * time.Millisecond)
			continue
		}

		defer resp.Body.Close()

		p.cb.ReportSuccess()

		var result []models.Book

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, appErr.ErrInvalidResponse
		}

		return result, nil
	}

	p.cb.ReportFailure()

	if errors.Is(lastErr, appErr.ErrTimeout) {
		return nil, appErr.ErrTimeout
	}

	return nil, errors.Join(appErr.ErrExternalService, lastErr)
}
