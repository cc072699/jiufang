// Package v1 implements the HTTP handlers for favorite management.
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

// FavoriteHandler handles favorite HTTP requests.
type FavoriteHandler struct {
	historyService *service.HistoryAppService
	logger         *zap.Logger
}

// NewFavoriteHandler creates a new FavoriteHandler instance.
func NewFavoriteHandler(historyService *service.HistoryAppService, logger *zap.Logger) *FavoriteHandler {
	return &FavoriteHandler{
		historyService: historyService,
		logger:         logger,
	}
}

// CreateFavoriteRequest represents the request body for creating a favorite.
type CreateFavoriteRequest struct {
	Name        string `json:"name" binding:"required"`
	Input       string `json:"input" binding:"required"`
	Sql         string `json:"sql" binding:"required"`
	Description string `json:"description"`
}

// FavoriteListResponse represents the response structure for favorite list.
type FavoriteListResponse struct {
	Favorites []FavoriteItem `json:"favorites"`
	Total     int64          `json:"total"`
	Page      int            `json:"page"`
	PageSize  int            `json:"page_size"`
}

// FavoriteItem represents a single favorite item in the response.
type FavoriteItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Input       string `json:"input"`
	Sql         string `json:"sql"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// CreateFavorite handles POST /api/v1/favorites - create a favorite.
func (h *FavoriteHandler) CreateFavorite(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse request body
	var req CreateFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// Create favorite
	favorite, err := h.historyService.CreateFavorite(ctx, userID, req.Name, req.Input, req.Sql, req.Description)
	if err != nil {
		h.logger.Error("Failed to create favorite",
			zap.Error(err),
			zap.String("name", req.Name),
		)
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Return success response
	response.Success(c, favorite)
}

// GetFavoriteList handles GET /api/v1/favorites - list favorites.
func (h *FavoriteHandler) GetFavoriteList(c *gin.Context) {
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
	name := c.Query("name")

	// Get favorite list
	favorites, total, err := h.historyService.GetFavoriteList(ctx, userID, page, pageSize, name)
	if err != nil {
		h.logger.Error("Failed to get favorite list", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "failed to get favorite list")
		return
	}

	// Convert favorites to FavoriteItem
	favoriteItems := make([]FavoriteItem, len(favorites))
	for i, fav := range favorites {
		favoriteItems[i] = FavoriteItem{
			ID:          strconv.FormatInt(fav.SnowflakeID, 10),
			Name:        fav.Name,
			Input:       fav.Input,
			Sql:         fav.Sql,
			Description: fav.Description,
			CreatedAt:   fav.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	// Create response with data structure
	data := FavoriteListResponse{
		Favorites: favoriteItems,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}

	// Return success response
	response.Success(c, data)
}

// DeleteFavorite handles DELETE /api/v1/favorites/:favorite_id - delete a favorite.
func (h *FavoriteHandler) DeleteFavorite(c *gin.Context) {
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

	// Delete favorite
	if err := h.historyService.DeleteFavorite(ctx, userID, favoriteID); err != nil {
		h.logger.Error("Failed to delete favorite",
			zap.Error(err),
			zap.String("favorite_id", favoriteID),
		)
		response.Error(c, http.StatusNotFound, "favorite not found")
		return
	}

	// Return success response
	response.Success(c, nil)
}
