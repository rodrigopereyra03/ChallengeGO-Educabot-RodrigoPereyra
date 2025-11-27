package services

import (
	"context"
	"slices"
	"strings"

	"educabot.com/bookshop/internal/books/models"
)

type MetricsService struct {
	booksProvider models.BooksProvider
}

func NewMetricsService(provider models.BooksProvider) *MetricsService {
	return &MetricsService{provider}
}

func (s *MetricsService) GetMetrics(ctx context.Context, author string) (map[string]interface{}, error) {
	books, err := s.booksProvider.GetBooks(ctx)
	if err != nil {
		return nil, err
	}

	mean := meanUnitsSold(books)
	cheapest := cheapestBook(books)
	written := booksWrittenByAuthor(books, author)

	return map[string]interface{}{
		"mean_units_sold":         mean,
		"cheapest_book":           cheapest.Name,
		"books_written_by_author": written,
	}, nil
}

func meanUnitsSold(books []models.Book) uint {
	if len(books) == 0 {
		return 0
	}

	var sum uint
	for _, b := range books {
		sum += b.UnitsSold
	}
	return sum / uint(len(books))
}

func cheapestBook(books []models.Book) models.Book {
	return slices.MinFunc(books, func(a, b models.Book) int {
		return int(a.Price - b.Price)
	})
}

func booksWrittenByAuthor(books []models.Book, author string) uint {
	var count uint

	normalized := strings.ToLower(strings.TrimSpace(author))

	for _, b := range books {
		if strings.ToLower(b.Author) == normalized {
			count++
		}
	}

	return count
}
