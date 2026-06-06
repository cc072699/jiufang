package service

import (
	"context"
	"time"

	"jiufang/internal/model/permission"
	pkgerrors "jiufang/internal/pkg/errors"
	"jiufang/internal/pkg/id"
	"jiufang/internal/repository"
)

type PermissionAppService struct {
	groupRepo      repository.UserGroupRepositoryInterface
	permissionRepo repository.PermissionRepositoryInterface
}

func NewPermissionAppService(groupRepo repository.UserGroupRepositoryInterface, permissionRepo repository.PermissionRepositoryInterface) *PermissionAppService {
	return &PermissionAppService{
		groupRepo:      groupRepo,
		permissionRepo: permissionRepo,
	}
}

type PermissionRequest struct {
	TableName       string `json:"table_name"`
	AllowedFields   string `json:"allowed_fields"`
	FilterCondition string `json:"filter_condition"`
}

type PermissionDetail struct {
	ID              uint
	SnowflakeID     int64
	GroupID         uint
	TableName       string
	AllowedFields   string
	FilterCondition string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (s *PermissionAppService) ConfigurePermissions(ctx context.Context, groupSnowflakeID int64, permissions []PermissionRequest) ([]permission.Permission, error) {
	group, err := s.groupRepo.GetBySnowflakeID(ctx, groupSnowflakeID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, pkgerrors.ErrGroupNotFound
	}

	newPermissions := make([]permission.Permission, 0)
	for _, req := range permissions {
		snowflakeID, err := id.Generate()
		if err != nil {
			return nil, err
		}

		p := permission.Permission{
			SnowflakeID:     snowflakeID,
			GroupID:         group.ID,
			TableName:       req.TableName,
			AllowedFields:   req.AllowedFields,
			FilterCondition: req.FilterCondition,
		}

		newPermissions = append(newPermissions, p)
	}

	if err := s.permissionRepo.ReplaceByGroupID(ctx, group.ID, newPermissions); err != nil {
		return nil, err
	}

	return newPermissions, nil
}

func (s *PermissionAppService) GetPermissionsByGroup(ctx context.Context, groupSnowflakeID int64) ([]PermissionDetail, error) {
	group, err := s.groupRepo.GetBySnowflakeID(ctx, groupSnowflakeID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, pkgerrors.ErrGroupNotFound
	}

	permissions, err := s.permissionRepo.GetByGroupID(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	permissionDetails := make([]PermissionDetail, 0)
	for _, p := range permissions {
		permissionDetails = append(permissionDetails, PermissionDetail{
			ID:              p.ID,
			SnowflakeID:     p.SnowflakeID,
			GroupID:         p.GroupID,
			TableName:       p.TableName,
			AllowedFields:   p.AllowedFields,
			FilterCondition: p.FilterCondition,
			CreatedAt:       p.CreatedAt,
			UpdatedAt:       p.UpdatedAt,
		})
	}

	return permissionDetails, nil
}
