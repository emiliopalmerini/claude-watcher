package limits

import (
	"fmt"
	"time"
)

// Service handles business logic for usage limit events
type Service struct {
	repo   Repository
	logger Logger
}

// NewService creates a new limits service
func NewService(repo Repository, logger Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// RecordLimitHit records a limit event with aggregate usage since the last limit
func (s *Service) RecordLimitHit(info ParsedLimitInfo) error {
	s.logger.Debug(fmt.Sprintf("Recording %s limit hit at %s", info.LimitType, info.Timestamp))

	// Get aggregate usage since last limit
	usage, err := s.repo.GetUsageSinceLastLimit()
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get usage since last limit: %v", err))
		return fmt.Errorf("get usage since last limit: %w", err)
	}

	// Create the limit event with aggregate data
	event := LimitEvent{
		Timestamp:      info.Timestamp,
		LimitType:      info.LimitType,
		SessionsCount:  usage.SessionsCount,
		InputTokens:    usage.InputTokens,
		OutputTokens:   usage.OutputTokens,
		ThinkingTokens: usage.ThinkingTokens,
		TotalCostUSD:   usage.TotalCostUSD,
	}

	// Save the limit event
	if err := s.repo.Save(event); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to save limit event: %v", err))
		return fmt.Errorf("save limit event: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("Recorded %s limit: %d sessions, $%.2f cost",
		info.LimitType, usage.SessionsCount, usage.TotalCostUSD))

	return nil
}

// RecordLimitHits records multiple limit events from a transcript
func (s *Service) RecordLimitHits(infos []ParsedLimitInfo) error {
	for _, info := range infos {
		if err := s.RecordLimitHit(info); err != nil {
			return err
		}
	}
	return nil
}

// GetRecentLimits returns limit events from the last N days
func (s *Service) GetRecentLimits(days int) ([]LimitEvent, error) {
	return s.repo.ListRecent(days)
}

// GetLastLimitTime returns when the last limit was hit
func (s *Service) GetLastLimitTime() (*time.Time, error) {
	return s.repo.GetLastLimitTimestamp()
}

// GetCurrentUsage returns usage metrics since the last limit event
func (s *Service) GetCurrentUsage() (UsageSummary, error) {
	return s.repo.GetUsageSinceLastLimit()
}
