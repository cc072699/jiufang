package service

import (
	"context"
	"time"

	"jiufang/internal/model/permission"
	pkgerrors "jiufang/internal/pkg/errors"
	"jiufang/internal/pkg/id"
	"jiufang/internal/repository"
)

type GroupAppService struct {
	groupRepo      repository.UserGroupRepositoryInterface
	userRepo       repository.UserRepositoryInterface
	permissionRepo repository.PermissionRepositoryInterface
}

func NewGroupAppService(groupRepo repository.UserGroupRepositoryInterface, userRepo repository.UserRepositoryInterface, permissionRepo repository.PermissionRepositoryInterface) *GroupAppService {
	return &GroupAppService{
		groupRepo:      groupRepo,
		userRepo:       userRepo,
		permissionRepo: permissionRepo,
	}
}

type CreateGroupRequest struct {
	Name        string
	Description string
	MemberIDs   []int64
}

type UpdateGroupRequest struct {
	Name        string
	Description string
	MemberIDs   []int64
}

type GroupDetail struct {
	ID          uint
	SnowflakeID int64
	Name        string
	Description string
	MemberCount int64
	Members     []GroupMember
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GroupMember struct {
	UserID      uint
	SnowflakeID int64
	Username    string
	Email       string
	Role        string
	CreatedAt   time.Time
}

func (s *GroupAppService) CreateGroup(ctx context.Context, req CreateGroupRequest) (*permission.UserGroup, error) {
	existingGroup, err := s.groupRepo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existingGroup != nil {
		return nil, pkgerrors.ErrGroupNameExists
	}

	snowflakeID, err := id.Generate()
	if err != nil {
		return nil, err
	}

	group := &permission.UserGroup{
		SnowflakeID: snowflakeID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	if len(req.MemberIDs) > 0 {
		userIDs := make([]uint, 0)
		for _, memberSnowflakeID := range req.MemberIDs {
			u, err := s.userRepo.GetBySnowflakeID(ctx, memberSnowflakeID)
			if err != nil {
				continue
			}
			if u != nil {
				userIDs = append(userIDs, u.ID)
			}
		}
		if len(userIDs) > 0 {
			if err := s.groupRepo.AddMembers(ctx, group.ID, userIDs); err != nil {
				return nil, err
			}
		}
	}

	return group, nil
}

func (s *GroupAppService) GetGroup(ctx context.Context, snowflakeID int64) (*GroupDetail, error) {
	group, err := s.groupRepo.GetBySnowflakeID(ctx, snowflakeID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, pkgerrors.ErrGroupNotFound
	}

	memberCount, err := s.groupRepo.GetMemberCount(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	members, err := s.groupRepo.GetMembers(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	groupMembers := make([]GroupMember, 0)
	for _, m := range members {
		u, err := s.userRepo.GetByID(ctx, m.UserID)
		if err != nil {
			continue
		}
		if u != nil {
			groupMembers = append(groupMembers, GroupMember{
				UserID:      u.ID,
				SnowflakeID: u.SnowflakeID,
				Username:    u.Username,
				Email:       u.Email,
				Role:        u.Role,
			})
		}
	}

	return &GroupDetail{
		ID:          group.ID,
		SnowflakeID: group.SnowflakeID,
		Name:        group.Name,
		Description: group.Description,
		MemberCount: memberCount,
		Members:     groupMembers,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}, nil
}

func (s *GroupAppService) ListGroups(ctx context.Context, page, pageSize int, name string) ([]GroupDetail, int64, error) {
	offset := (page - 1) * pageSize
	groups, total, err := s.groupRepo.List(ctx, offset, pageSize, name)
	if err != nil {
		return nil, 0, err
	}

	groupDetails := make([]GroupDetail, 0)
	for _, group := range groups {
		memberCount, err := s.groupRepo.GetMemberCount(ctx, group.ID)
		if err != nil {
			memberCount = 0
		}

		groupDetails = append(groupDetails, GroupDetail{
			ID:          group.ID,
			SnowflakeID: group.SnowflakeID,
			Name:        group.Name,
			Description: group.Description,
			MemberCount: memberCount,
			CreatedAt:   group.CreatedAt,
			UpdatedAt:   group.UpdatedAt,
		})
	}

	return groupDetails, total, nil
}

func (s *GroupAppService) UpdateGroup(ctx context.Context, snowflakeID int64, req UpdateGroupRequest) (*permission.UserGroup, error) {
	group, err := s.groupRepo.GetBySnowflakeID(ctx, snowflakeID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, pkgerrors.ErrGroupNotFound
	}

	if req.Name != "" && req.Name != group.Name {
		existingGroup, err := s.groupRepo.GetByName(ctx, req.Name)
		if err != nil {
			return nil, err
		}
		if existingGroup != nil {
			return nil, pkgerrors.ErrGroupNameExists
		}
		group.Name = req.Name
	}

	if req.Description != "" {
		group.Description = req.Description
	}

	if err := s.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	if req.MemberIDs != nil {
		existingMembers, err := s.groupRepo.GetMembers(ctx, group.ID)
		if err != nil {
			return nil, err
		}

		existingUserIDs := make(map[uint]bool)
		for _, m := range existingMembers {
			existingUserIDs[m.UserID] = true
		}

		newUserIDs := make(map[uint]bool)
		userIDsToAdd := make([]uint, 0)
		userIDsToRemove := make([]uint, 0)

		for _, memberSnowflakeID := range req.MemberIDs {
			u, err := s.userRepo.GetBySnowflakeID(ctx, memberSnowflakeID)
			if err != nil {
				continue
			}
			if u != nil {
				newUserIDs[u.ID] = true
				if !existingUserIDs[u.ID] {
					userIDsToAdd = append(userIDsToAdd, u.ID)
				}
			}
		}

		for userID := range existingUserIDs {
			if !newUserIDs[userID] {
				userIDsToRemove = append(userIDsToRemove, userID)
			}
		}

		if len(userIDsToAdd) > 0 {
			if err := s.groupRepo.AddMembers(ctx, group.ID, userIDsToAdd); err != nil {
				return nil, err
			}
		}

		if len(userIDsToRemove) > 0 {
			if err := s.groupRepo.RemoveMembers(ctx, group.ID, userIDsToRemove); err != nil {
				return nil, err
			}
		}
	}

	return group, nil
}

func (s *GroupAppService) DeleteGroup(ctx context.Context, snowflakeID int64) error {
	group, err := s.groupRepo.GetBySnowflakeID(ctx, snowflakeID)
	if err != nil {
		return err
	}
	if group == nil {
		return pkgerrors.ErrGroupNotFound
	}

	isPreset, err := s.groupRepo.IsPresetGroup(ctx, snowflakeID)
	if err != nil {
		return err
	}
	if isPreset {
		return pkgerrors.ErrPresetGroupCannotDelete
	}

	return s.groupRepo.Delete(ctx, group.ID)
}

// GetGroupMembers retrieves group members with pagination.
func (s *GroupAppService) GetGroupMembers(ctx context.Context, groupSnowflakeID int64, page, pageSize int) ([]GroupMember, int64, error) {
	// Validate and set default pagination
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// Get group by snowflake ID
	group, err := s.groupRepo.GetBySnowflakeID(ctx, groupSnowflakeID)
	if err != nil {
		return nil, 0, err
	}
	if group == nil {
		return nil, 0, pkgerrors.ErrGroupNotFound
	}

	// Get members with pagination
	offset := (page - 1) * pageSize
	members, total, err := s.groupRepo.GetMembersWithPagination(ctx, group.ID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// Build group member details
	groupMembers := make([]GroupMember, 0)
	for _, m := range members {
		u, err := s.userRepo.GetByID(ctx, m.UserID)
		if err != nil {
			continue
		}
		if u != nil {
			groupMembers = append(groupMembers, GroupMember{
				UserID:      u.ID,
				SnowflakeID: u.SnowflakeID,
				Username:    u.Username,
				Role:        u.Role,
				Email:       u.Email,
				CreatedAt:   m.CreatedAt,
			})
		}
	}

	return groupMembers, total, nil
}

// AddGroupMembers adds multiple members to a group.
func (s *GroupAppService) AddGroupMembers(ctx context.Context, groupSnowflakeID int64, userSnowflakeIDs []int64) ([]GroupMember, error) {
	// Get group by snowflake ID
	group, err := s.groupRepo.GetBySnowflakeID(ctx, groupSnowflakeID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, pkgerrors.ErrGroupNotFound
	}

	// Validate and collect user IDs
	userIDs := make([]uint, 0)
	groupMembers := make([]GroupMember, 0)

	for _, userSnowflakeID := range userSnowflakeIDs {
		u, err := s.userRepo.GetBySnowflakeID(ctx, userSnowflakeID)
		if err != nil {
			continue
		}
		if u != nil {
			userIDs = append(userIDs, u.ID)
			groupMembers = append(groupMembers, GroupMember{
				UserID:      u.ID,
				SnowflakeID: u.SnowflakeID,
				Username:    u.Username,
				Role:        u.Role,
				Email:       u.Email,
			})
		}
	}

	if len(userIDs) == 0 {
		return nil, pkgerrors.ErrInvalidRequest
	}

	// Add members to group
	if err := s.groupRepo.AddMembers(ctx, group.ID, userIDs); err != nil {
		return nil, err
	}

	return groupMembers, nil
}

// RemoveGroupMember removes a single member from a group.
func (s *GroupAppService) RemoveGroupMember(ctx context.Context, groupSnowflakeID int64, userSnowflakeID int64) error {
	// Get group by snowflake ID
	group, err := s.groupRepo.GetBySnowflakeID(ctx, groupSnowflakeID)
	if err != nil {
		return err
	}
	if group == nil {
		return pkgerrors.ErrGroupNotFound
	}

	// Get user by snowflake ID
	user, err := s.userRepo.GetBySnowflakeID(ctx, userSnowflakeID)
	if err != nil {
		return err
	}
	if user == nil {
		return pkgerrors.ErrUserNotFound
	}

	// Remove member from group
	if err := s.groupRepo.RemoveMembers(ctx, group.ID, []uint{user.ID}); err != nil {
		return err
	}

	return nil
}
