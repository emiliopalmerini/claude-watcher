package ports

import (
	"context"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

type UsageLimitsRepository interface {
	Upsert(ctx context.Context, limit *domain.UsageLimit) error
	Get(ctx context.Context, id string) (*domain.UsageLimit, error)
	List(ctx context.Context) ([]*domain.UsageLimit, error)
	Delete(ctx context.Context, id string) error
}
