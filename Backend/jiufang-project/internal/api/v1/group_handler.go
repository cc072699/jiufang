package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	pkgerrors "jiufang/internal/pkg/errors"
	"jiufang/internal/pkg/response"
	"jiufang/internal/service"
)

type GroupHandler struct {
	groupService      *service.GroupAppService
	permissionService *service.PermissionAppService
}

func NewGroupHandler(groupService *service.GroupAppService, permissionService *service.PermissionAppService) *GroupHandler {
	return &GroupHandler{
		groupService:      groupService,
		permissionService: permissionService,
	}
}

type CreateGroupRequest struct {
	Name        string  `json:"name" binding:"required,min=3,max=50"`
	Description string  `json:"description" binding:"max=200"`
	Members     []int64 `json:"members"`
}

type UpdateGroupRequest struct {
	Name        string  `json:"name" binding:"min=3,max=50"`
	Description string  `json:"description" binding:"max=200"`
	Members     []int64 `json:"members"`
}

type ConfigurePermissionsRequest struct {
	Permissions []service.PermissionRequest `json:"permissions" binding:"required"`
}

// GroupResponse represents the response data for user group operations
type GroupResponse struct {
	ID          string `json:"id"` // Snowflake ID as string to avoid JavaScript precision loss
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

// GroupListResponse represents the response data for listing user groups
type GroupListResponse struct {
	ID          string `json:"id"` // Snowflake ID as string
	Name        string `json:"name"`
	Description string `json:"description"`
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	group, err := h.groupService.CreateGroup(c.Request.Context(), service.CreateGroupRequest{
		Name:        req.Name,
		Description: req.Description,
		MemberIDs:   req.Members,
	})
	if err != nil {
		if err == pkgerrors.ErrGroupNameExists {
			response.Error(c, http.StatusConflict, "user group name already exists")
			return
		}
		response.InternalError(c, "failed to create user group")
		return
	}

	// Fetch members for the newly created group
	members := make([]gin.H, 0)
	fullGroup, err := h.groupService.GetGroup(c.Request.Context(), group.SnowflakeID)
	if err == nil && fullGroup.Members != nil {
		for _, m := range fullGroup.Members {
			members = append(members, gin.H{
				"id":       strconv.FormatInt(m.SnowflakeID, 10),
				"username": m.Username,
				"email":    m.Email,
				"role":     m.Role,
			})
		}
	}

	groupResp := gin.H{
		"id":          strconv.FormatInt(group.SnowflakeID, 10),
		"name":        group.Name,
		"description": group.Description,
		"members":     members,
		"created_at":  group.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	response.Success(c, groupResp)
}

func (h *GroupHandler) ListGroups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	name := c.Query("name")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	groups, total, err := h.groupService.ListGroups(c.Request.Context(), page, pageSize, name)
	if err != nil {
		response.InternalError(c, "failed to list user groups")
		return
	}

	groupList := make([]gin.H, 0)
	for _, group := range groups {
		groupList = append(groupList, gin.H{
			"id":           strconv.FormatInt(group.SnowflakeID, 10),
			"name":         group.Name,
			"description":  group.Description,
			"member_count": group.MemberCount,
			"created_at":   group.CreatedAt,
		})
	}

	response.Success(c, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"groups":    groupList,
	})
}

func (h *GroupHandler) GetGroup(c *gin.Context) {
	groupIDStr := c.Param("group_id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid group_id format")
		return
	}

	group, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err != nil {
		if err == pkgerrors.ErrGroupNotFound {
			response.NotFound(c, "user group not found")
			return
		}
		response.InternalError(c, "failed to get user group")
		return
	}

	members := make([]gin.H, 0)
	for _, m := range group.Members {
		members = append(members, gin.H{
			"id":       strconv.FormatInt(m.SnowflakeID, 10),
			"username": m.Username,
			"email":    m.Email,
			"role":     m.Role,
		})
	}

	response.Success(c, gin.H{
		"id":           strconv.FormatInt(group.SnowflakeID, 10),
		"name":         group.Name,
		"description":  group.Description,
		"member_count": group.MemberCount,
		"members":      members,
		"created_at":   group.CreatedAt,
		"updated_at":   group.UpdatedAt,
	})
}

func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	groupIDStr := c.Param("group_id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid group_id format")
		return
	}

	var req UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	group, err := h.groupService.UpdateGroup(c.Request.Context(), groupID, service.UpdateGroupRequest{
		Name:        req.Name,
		Description: req.Description,
		MemberIDs:   req.Members,
	})
	if err != nil {
		if err == pkgerrors.ErrGroupNotFound {
			response.NotFound(c, "user group not found")
			return
		}
		if err == pkgerrors.ErrGroupNameExists {
			response.Error(c, http.StatusConflict, "user group name already exists")
			return
		}
		response.InternalError(c, "failed to update user group")
		return
	}

	// Fetch members for the updated group
	members := make([]gin.H, 0)
	fullGroup, err := h.groupService.GetGroup(c.Request.Context(), groupID)
	if err == nil && fullGroup.Members != nil {
		for _, m := range fullGroup.Members {
			members = append(members, gin.H{
				"id":       strconv.FormatInt(m.SnowflakeID, 10),
				"username": m.Username,
				"email":    m.Email,
				"role":     m.Role,
			})
		}
	}

	response.Success(c, gin.H{
		"id":          strconv.FormatInt(group.SnowflakeID, 10),
		"name":        group.Name,
		"description": group.Description,
		"members":     members,
		"updated_at":  group.UpdatedAt,
	})
}

func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	groupIDStr := c.Param("group_id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid group_id format")
		return
	}

	err = h.groupService.DeleteGroup(c.Request.Context(), groupID)
	if err != nil {
		if err == pkgerrors.ErrGroupNotFound {
			response.NotFound(c, "user group not found")
			return
		}
		if err == pkgerrors.ErrPresetGroupCannotDelete {
			response.Error(c, http.StatusForbidden, "preset user group cannot be deleted")
			return
		}
		response.InternalError(c, "failed to delete user group")
		return
	}

	response.Success(c, nil)
}

func (h *GroupHandler) ConfigurePermissions(c *gin.Context) {
	groupIDStr := c.Param("group_id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid group_id format")
		return
	}

	var req ConfigurePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	permissions, err := h.permissionService.ConfigurePermissions(c.Request.Context(), groupID, req.Permissions)
	if err != nil {
		if err == pkgerrors.ErrGroupNotFound {
			response.NotFound(c, "user group not found")
			return
		}
		response.InternalError(c, "failed to configure permissions")
		return
	}

	permissionList := make([]gin.H, 0)
	for _, p := range permissions {
		permissionList = append(permissionList, gin.H{
			"id":               strconv.FormatInt(p.SnowflakeID, 10),
			"table_name":       p.TableName,
			"allowed_fields":   p.AllowedFields,
			"filter_condition": p.FilterCondition,
		})
	}

	response.Success(c, gin.H{
		"group_id":    strconv.FormatInt(groupID, 10),
		"permissions": permissionList,
		"created_at":  time.Now().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *GroupHandler) GetPermissions(c *gin.Context) {
	groupIDStr := c.Param("group_id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid group_id format")
		return
	}

	permissions, err := h.permissionService.GetPermissionsByGroup(c.Request.Context(), groupID)
	if err != nil {
		if err == pkgerrors.ErrGroupNotFound {
			response.NotFound(c, "user group not found")
			return
		}
		response.InternalError(c, "failed to get permissions")
		return
	}

	permissionList := make([]gin.H, 0)
	for _, p := range permissions {
		permissionList = append(permissionList, gin.H{
			"id":               strconv.FormatInt(p.SnowflakeID, 10),
			"table_name":       p.TableName,
			"allowed_fields":   p.AllowedFields,
			"filter_condition": p.FilterCondition,
		})
	}

	response.Success(c, gin.H{
		"group_id":    strconv.FormatInt(groupID, 10),
		"permissions": permissionList,
	})
}

type AddGroupMembersRequest struct {
	UserID  string  `json:"user_id"`  // Single user ID (string, from frontend)
	UserIDs []int64 `json:"user_ids"` // Multiple user IDs (legacy)
}

func (h *GroupHandler) GetGroupMembers(c *gin.Context) {
	groupIDStr := c.Param("group_id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid group_id format")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	members, total, err := h.groupService.GetGroupMembers(c.Request.Context(), groupID, page, pageSize)
	if err != nil {
		if err == pkgerrors.ErrGroupNotFound {
			response.NotFound(c, "user group not found")
			return
		}
		response.InternalError(c, "failed to get group members")
		return
	}

	memberList := make([]gin.H, 0)
	for _, m := range members {
		memberList = append(memberList, gin.H{
			"id":         strconv.FormatInt(m.SnowflakeID, 10),
			"username":   m.Username,
			"email":      m.Email,
			"role":       m.Role,
			"created_at": m.CreatedAt,
		})
	}

	response.Success(c, gin.H{
		"members":   memberList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *GroupHandler) AddGroupMembers(c *gin.Context) {
	groupIDStr := c.Param("group_id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid group_id format")
		return
	}

	var req AddGroupMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request parameters")
		return
	}

	// Resolve user IDs from either single user_id or user_ids array
	var userIDs []int64
	if req.UserID != "" {
		id, err := strconv.ParseInt(req.UserID, 10, 64)
		if err != nil {
			response.BadRequest(c, "invalid user_id format")
			return
		}
		userIDs = []int64{id}
	} else if len(req.UserIDs) > 0 {
		userIDs = req.UserIDs
	} else {
		response.BadRequest(c, "user_id or user_ids is required")
		return
	}

	members, err := h.groupService.AddGroupMembers(c.Request.Context(), groupID, userIDs)
	if err != nil {
		if err == pkgerrors.ErrGroupNotFound {
			response.NotFound(c, "user group not found")
			return
		}
		if err == pkgerrors.ErrInvalidRequest {
			response.BadRequest(c, "invalid user IDs")
			return
		}
		response.InternalError(c, "failed to add group members")
		return
	}

	memberList := make([]gin.H, 0)
	for _, m := range members {
		memberList = append(memberList, gin.H{
			"id":       strconv.FormatInt(m.SnowflakeID, 10),
			"username": m.Username,
			"email":    m.Email,
			"role":     m.Role,
		})
	}

	response.Success(c, gin.H{
		"added_count": len(members),
		"members":     memberList,
	})
}

func (h *GroupHandler) RemoveGroupMember(c *gin.Context) {
	groupIDStr := c.Param("group_id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid group_id format")
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user_id format")
		return
	}

	err = h.groupService.RemoveGroupMember(c.Request.Context(), groupID, userID)
	if err != nil {
		if err == pkgerrors.ErrGroupNotFound {
			response.NotFound(c, "user group not found")
			return
		}
		if err == pkgerrors.ErrUserNotFound {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalError(c, "failed to remove group member")
		return
	}

	response.Success(c, nil)
}
