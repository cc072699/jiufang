// Package service implements the application layer for feedback management.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"jiufang/internal/model/feedback"
	"jiufang/internal/pkg/id"
	"jiufang/internal/repository"
)

// FeedbackServiceInterface defines the interface for feedback business service.
type FeedbackServiceInterface interface {
	CreateFeedback(ctx context.Context, userID int64, req *feedback.CreateFeedbackRequest) (*feedback.Feedback, error)
	GetFeedbackByID(ctx context.Context, snowflakeID int64) (*feedback.Feedback, error)
	ListFeedbacks(ctx context.Context, req *feedback.ListFeedbacksRequest) ([]feedback.Feedback, int64, error)
}

// FeedbackService manages feedback business logic.
type FeedbackService struct {
	feedbackRepo repository.FeedbackRepositoryInterface
	queryRepo    repository.QueryRepositoryInterface
	idGenerator  id.SnowflakeGeneratorInterface
	logger       *zap.Logger
}

// NewFeedbackService creates a new FeedbackService instance.
func NewFeedbackService(
	feedbackRepo repository.FeedbackRepositoryInterface,
	queryRepo repository.QueryRepositoryInterface,
	idGenerator id.SnowflakeGeneratorInterface,
	logger *zap.Logger,
) *FeedbackService {
	return &FeedbackService{
		feedbackRepo: feedbackRepo,
		queryRepo:    queryRepo,
		idGenerator:  idGenerator,
		logger:       logger,
	}
}

// CreateFeedback creates a new feedback.
func (s *FeedbackService) CreateFeedback(ctx context.Context, userID int64, req *feedback.CreateFeedbackRequest) (*feedback.Feedback, error) {
	// Validate rating value
	if req.Rating != "satisfied" && req.Rating != "unsatisfied" {
		return nil, errors.New("invalid rating value, must be 'satisfied' or 'unsatisfied'")
	}

	// Validate reason is required when rating is unsatisfied
	if req.Rating == "unsatisfied" && req.Reason == "" {
		return nil, errors.New("reason is required when rating is unsatisfied")
	}

	// Validate reason length
	if len(req.Reason) > 500 {
		return nil, errors.New("reason length must not exceed 500 characters")
	}

	// Check if feedback already exists for this query record
	exists, err := s.feedbackRepo.IsFeedbackExists(ctx, req.QueryRecordID)
	if err != nil {
		s.logger.Error("failed to check feedback existence",
			zap.Int64("query_record_id", req.QueryRecordID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check feedback existence: %w", err)
	}
	if exists {
		return nil, errors.New("feedback already exists for this query record")
	}

	// Get query record to retrieve query question
	queryRecord, err := s.queryRepo.GetQueryRecordBySnowflakeID(ctx, req.QueryRecordID)
	if err != nil {
		s.logger.Error("failed to get query record",
			zap.Int64("query_record_id", req.QueryRecordID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get query record: %w", err)
	}

	// Verify the query record belongs to the user
	if queryRecord.UserID != userID {
		return nil, errors.New("query record does not belong to the user")
	}

	// Generate snowflake ID
	snowflakeID := s.idGenerator.Generate()

	// Create feedback entity
	fb := &feedback.Feedback{
		SnowflakeID:   snowflakeID,
		UserID:        userID,
		QueryRecordID: req.QueryRecordID,
		QueryQuestion: queryRecord.Input,
		Rating:        feedback.Rating(req.Rating),
		Reason:        req.Reason,
		CreatedAt:     time.Now(),
	}

	// Save to database
	if err := s.feedbackRepo.Create(ctx, fb); err != nil {
		s.logger.Error("failed to create feedback",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Int64("user_id", userID),
			zap.Int64("query_record_id", req.QueryRecordID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create feedback: %w", err)
	}

	s.logger.Info("feedback created successfully",
		zap.Int64("snowflake_id", snowflakeID),
		zap.Int64("user_id", userID),
		zap.String("rating", req.Rating),
	)

	return fb, nil
}

// GetFeedbackByID retrieves a feedback by its snowflake ID.
func (s *FeedbackService) GetFeedbackByID(ctx context.Context, snowflakeID int64) (*feedback.Feedback, error) {
	fb, err := s.feedbackRepo.GetBySnowflakeID(ctx, snowflakeID)
	if err != nil {
		s.logger.Error("failed to get feedback",
			zap.Int64("snowflake_id", snowflakeID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get feedback: %w", err)
	}

	return fb, nil
}

// ListFeedbacks retrieves feedbacks with pagination and filters.
func (s *FeedbackService) ListFeedbacks(ctx context.Context, req *feedback.ListFeedbacksRequest) ([]feedback.Feedback, int64, error) {
	// Set default pagination values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 10
	}
	if req.Size > 100 {
		req.Size = 100
	}

	// Calculate offset
	offset := (req.Page - 1) * req.Size

	// Get feedbacks from repository
	feedbacks, total, err := s.feedbackRepo.List(ctx, offset, req.Size, req.UserID, req.Rating)
	if err != nil {
		s.logger.Error("failed to list feedbacks",
			zap.Int("page", req.Page),
			zap.Int("size", req.Size),
			zap.Int64("user_id", req.UserID),
			zap.String("rating", req.Rating),
			zap.Error(err),
		)
		return nil, 0, fmt.Errorf("failed to list feedbacks: %w", err)
	}

	s.logger.Info("feedbacks listed successfully",
		zap.Int("page", req.Page),
		zap.Int("size", req.Size),
		zap.Int64("total", total),
	)

	return feedbacks, total, nil
}