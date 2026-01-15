package analytics

import "context"

// Service provides analytics business logic
type Service struct {
	repo   Repository
	logger Logger
}

// NewService creates a new analytics service
func NewService(repo Repository, logger Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// GetOverview returns aggregate metrics for the dashboard overview
func (s *Service) GetOverview(ctx context.Context) (OverviewMetrics, error) {
	s.logger.Debug("Fetching overview metrics")
	return s.repo.GetOverviewMetrics(ctx)
}
