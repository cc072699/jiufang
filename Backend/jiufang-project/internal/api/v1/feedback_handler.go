package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"jiufang/internal/model/feedback"
	"jiufang/internal/pkg/response"
	"jiufang/internal/service"
)

// FeedbackHandler handles HTTP requests for feedbacks.
type FeedbackHandler struct {
	feedbackService service.FeedbackServiceInterface
	logger          *zap.Logger
}

// NewFeedbackHandler creates a new FeedbackHandler instance.
func NewFeedbackHandler(
	feedbackService service.FeedbackServiceInterface,
	logger *zap.Logger,
) *FeedbackHandler {
	return &FeedbackHandler{
		feedbackService: feedbackService,
		logger:          logger,
	}
}

// CreateFeedback handles POST /api/v1/feedbacks requests.
func (h *FeedbackHandler) CreateFeedback(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req feedback.CreateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters: "+err.Error())
		return
	}

	// Create feedback
	fb, err := h.feedbackService.CreateFeedback(ctx, userID.(int64), &req)
	if err != nil {
		h.logger.Error("failed to create feedback",
			zap.Int64("user_id", userID.(int64)),
			zap.Int64("query_record_id", req.QueryRecordID),
			zap.Error(err),
		)
		response.InternalError(c, "failed to create feedback: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":              fb.SnowflakeID,
		"user_id":         fb.UserID,
		"query_record_id": fb.QueryRecordID,
		"query_question":  fb.QueryQuestion,
		"rating":          fb.Rating,
		"reason":          fb.Reason,
		"created_at":      fb.CreatedAt,
	})
}

// GetFeedback handles GET /api/v1/feedbacks/:id requests.
func (h *FeedbackHandler) GetFeedback(c *gin.Context) {
	ctx := c.Request.Context()

	// Get feedback ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid feedback id")
		return
	}

	// Get feedback
	fb, err := h.feedbackService.GetFeedbackByID(ctx, id)
	if err != nil {
		h.logger.Error("failed to get feedback",
			zap.Int64("id", id),
			zap.Error(err),
		)
		response.InternalError(c, "failed to get feedback: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":              fb.SnowflakeID,
		"user_id":         fb.UserID,
		"query_record_id": fb.QueryRecordID,
		"query_question":  fb.QueryQuestion,
		"rating":          fb.Rating,
		"reason":          fb.Reason,
		"created_at":      fb.CreatedAt,
	})
}

// ListFeedbacks handles GET /api/v1/feedbacks requests.
func (h *FeedbackHandler) ListFeedbacks(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// Parse filter parameters
	userIDStr := c.Query("user_id")
	var userID int64
	if userIDStr != "" {
		userID, _ = strconv.ParseInt(userIDStr, 10, 64)
	}

	rating := c.Query("rating")

	// Build request
	req := &feedback.ListFeedbacksRequest{
		Page:   page,
		Size:   pageSize,
		UserID: userID,
		Rating: rating,
	}

	// List feedbacks
	feedbacks, total, err := h.feedbackService.ListFeedbacks(ctx, req)
	if err != nil {
		h.logger.Error("failed to list feedbacks",
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
			zap.Error(err),
		)
		response.InternalError(c, "failed to list feedbacks: "+err.Error())
		return
	}

	// Build response list
	var list []gin.H
	for _, fb := range feedbacks {
		list = append(list, gin.H{
			"id":              fb.SnowflakeID,
			"user_id":         fb.UserID,
			"query_record_id": fb.QueryRecordID,
			"query_question":  fb.QueryQuestion,
			"rating":          fb.Rating,
			"reason":          fb.Reason,
			"created_at":      fb.CreatedAt,
		})
	}

	response.PageWithField(c, "feedbacks", list, total, page, pageSize)
}
