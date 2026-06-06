// Package v1 implements the HTTP handlers for API version 1.
package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"jiufang/internal/middleware"
	"jiufang/internal/pkg/response"
	"jiufang/internal/service"
)

// DialogFavoriteHandler handles dialog favorite related HTTP requests.
type DialogFavoriteHandler struct {
	dialogFavoriteService *service.DialogFavoriteAppService
	logger                *zap.Logger
}

// NewDialogFavoriteHandler creates a new DialogFavoriteHandler instance.
func NewDialogFavoriteHandler(dialogFavoriteService *service.DialogFavoriteAppService, logger *zap.Logger) *DialogFavoriteHandler {
	return &DialogFavoriteHandler{
		dialogFavoriteService: dialogFavoriteService,
		logger:                logger,
	}
}

// CreateDialogFavoriteRequest represents the request body for creating a dialog favorite.
type CreateDialogFavoriteRequest struct {
	DialogSessionID string `json:"dialog_session_id" binding:"required"`
	Title           string `json:"title"`
}

// CreateDialogFavorite handles POST /api/v1/dialog-favorites - create a dialog favorite.
func (h *DialogFavoriteHandler) CreateDialogFavorite(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse request body
	var req CreateDialogFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// Create dialog favorite
	favorite, err := h.dialogFavoriteService.CreateDialogFavorite(ctx, userID, req.DialogSessionID, req.Title)
	if err != nil {
		h.logger.Error("Failed to create dialog favorite",
			zap.Error(err),
			zap.String("dialog_session_id", req.DialogSessionID),
		)
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Return success response
	response.Success(c, favorite)
}

// GetDialogFavoriteList handles GET /api/v1/dialog-favorites - list dialog favorites.
func (h *DialogFavoriteHandler) GetDialogFavoriteList(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Get dialog favorite list
	favorites, total, err := h.dialogFavoriteService.GetDialogFavoriteList(ctx, userID, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to get dialog favorite list", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get dialog favorite list")
		return
	}

	// Return paginated response
	response.PageWithField(c, "favorites", favorites, total, page, pageSize)
}

// DeleteDialogFavorite handles DELETE /api/v1/dialog-favorites/:favorite_id - delete a dialog favorite.
func (h *DialogFavoriteHandler) DeleteDialogFavorite(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get favorite ID from path parameter
	favoriteID := c.Param("favorite_id")
	if favoriteID == "" {
		response.Error(c, http.StatusBadRequest, "favorite_id is required")
		return
	}

	// Delete dialog favorite
	if err := h.dialogFavoriteService.DeleteDialogFavorite(ctx, userID, favoriteID); err != nil {
		h.logger.Error("Failed to delete dialog favorite",
			zap.Error(err),
			zap.String("favorite_id", favoriteID),
		)
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Return success response
	response.Success(c, nil)
}
