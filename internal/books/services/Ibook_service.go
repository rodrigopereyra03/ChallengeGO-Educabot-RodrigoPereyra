package services

import "context"

type MetricsServiceI interface {
	GetMetrics(ctx context.Context, author string) (map[string]interface{}, error)
}
