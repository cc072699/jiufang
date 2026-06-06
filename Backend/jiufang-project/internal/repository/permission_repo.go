package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"jiufang/internal/model/permission"
)

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) Create(ctx context.Context, p *permission.Permission) error {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}
	return nil
}

func (r *PermissionRepository) CreateBatch(ctx context.Context, permissions []permission.Permission) error {
	if err := r.db.WithContext(ctx).Create(&permissions).Error; err != nil {
		return fmt.Errorf("failed to create permissions batch: %w", err)
	}
	return nil
}

func (r *PermissionRepository) GetByID(ctx context.Context, id uint) (*permission.Permission, error) {
	var p permission.Permission
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get permission by id: %w", err)
	}
	return &p, nil
}

func (r *PermissionRepository) GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*permission.Permission, error) {
	var p permission.Permission
	err := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get permission by snowflake_id: %w", err)
	}
	return &p, nil
}

func (r *PermissionRepository) GetByGroupID(ctx context.Context, groupID uint) ([]permission.Permission, error) {
	var permissions []permission.Permission
	err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Find(&permissions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions by group_id: %w", err)
	}
	return permissions, nil
}

func (r *PermissionRepository) Update(ctx context.Context, p *permission.Permission) error {
	result := r.db.WithContext(ctx).Save(p)
	if result.Error != nil {
		return fmt.Errorf("failed to update permission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("permission not found")
	}
	return nil
}

func (r *PermissionRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&permission.Permission{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete permission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("permission not found")
	}
	return nil
}

func (r *PermissionRepository) DeleteByGroupID(ctx context.Context, groupID uint) error {
	result := r.db.WithContext(ctx).Where("group_id = ?", groupID).Delete(&permission.Permission{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete permissions by group_id: %w", result.Error)
	}
	return nil
}

func (r *PermissionRepository) List(ctx context.Context, offset, limit int, groupID uint, resourceType string) ([]permission.Permission, int64, error) {
	var permissions []permission.Permission
	var total int64

	query := r.db.WithContext(ctx).Model(&permission.Permission{})

	if groupID > 0 {
		query = query.Where("group_id = ?", groupID)
	}
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count permissions: %w", err)
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&permissions).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list permissions: %w", err)
	}

	return permissions, total, nil
}