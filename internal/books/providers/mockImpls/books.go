package mockImpls

import (
	"context"

	"educabot.com/bookshop/internal/books/models"
)

type MockBooksProvider struct{}

func NewMockBooksProvider() *MockBooksProvider {
	return &MockBooksProvider{}
}

func (m *MockBooksProvider) GetBooks(ctx context.Context) ([]models.Book, error) {
	return []models.Book{
		{ID: 1, Name: "The Go Programming Language", Author: "Alan Donovan", UnitsSold: 5000, Price: 40},
		{ID: 2, Name: "Clean Code", Author: "Robert C. Martin", UnitsSold: 15000, Price: 50},
		{ID: 3, Name: "The Pragmatic Programmer", Author: "Andrew Hunt", UnitsSold: 13000, Price: 45},
	}, nil
}
