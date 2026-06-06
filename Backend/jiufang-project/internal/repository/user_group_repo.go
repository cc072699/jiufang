package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"jiufang/internal/model/permission"
	"jiufang/internal/model/user"
)

type UserGroupRepository struct {
	db *gorm.DB
}

func NewUserGroupRepository(db *gorm.DB) *UserGroupRepository {
	return &UserGroupRepository{db: db}
}

func (r *UserGroupRepository) Create(ctx context.Context, group *permission.UserGroup) error {
	if err := r.db.WithContext(ctx).Create(group).Error; err != nil {
		return fmt.Errorf("failed to create user group: %w", err)
	}
	return nil
}

func (r *UserGroupRepository) GetByID(ctx context.Context, id uint) (*permission.UserGroup, error) {
	var group permission.UserGroup
	err := r.db.WithContext(ctx).First(&group, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user group by id: %w", err)
	}
	return &group, nil
}

func (r *UserGroupRepository) GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*permission.UserGroup, error) {
	var group permission.UserGroup
	err := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).First(&group).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user group by snowflake_id: %w", err)
	}
	return &group, nil
}

func (r *UserGroupRepository) GetByName(ctx context.Context, name string) (*permission.UserGroup, error) {
	var group permission.UserGroup
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&group).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user group by name: %w", err)
	}
	return &group, nil
}

func (r *UserGroupRepository) List(ctx context.Context, offset, limit int, name string) ([]permission.UserGroup, int64, error) {
	var groups []permission.UserGroup
	var total int64

	query := r.db.WithContext(ctx).Model(&permission.UserGroup{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user groups: %w", err)
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&groups).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list user groups: %w", err)
	}

	return groups, total, nil
}

func (r *UserGroupRepository) Update(ctx context.Context, group *permission.UserGroup) error {
	result := r.db.WithContext(ctx).Save(group)
	if result.Error != nil {
		return fmt.Errorf("failed to update user group: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("user group not found")
	}
	return nil
}

func (r *UserGroupRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&permission.UserGroup{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user group: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("user group not found")
	}
	return nil
}

func (r *UserGroupRepository) GetMemberCount(ctx context.Context, groupID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&user.UserGroupMember{}).Where("group_id = ?", groupID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count group members: %w", err)
	}
	return count, nil
}

func (r *UserGroupRepository) AddMembers(ctx context.Context, groupID uint, userIDs []uint) error {
	members := make([]user.UserGroupMember, len(userIDs))
	for i, userID := range userIDs {
		members[i] = user.UserGroupMember{
			UserID:  userID,
			GroupID: groupID,
		}
	}

	if err := r.db.WithContext(ctx).Create(&members).Error; err != nil {
		return fmt.Errorf("failed to add group members: %w", err)
	}
	return nil
}

func (r *UserGroupRepository) RemoveMembers(ctx context.Context, groupID uint, userIDs []uint) error {
	result := r.db.WithContext(ctx).Where("group_id = ? AND user_id IN ?", groupID, userIDs).Delete(&user.UserGroupMember{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove group members: %w", result.Error)
	}
	return nil
}

func (r *UserGroupRepository) GetMembers(ctx context.Context, groupID uint) ([]user.UserGroupMember, error) {
	var members []user.UserGroupMember
	err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Find(&members).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	return members, nil
}

func (r *UserGroupRepository) GetMembersWithPagination(ctx context.Context, groupID uint, offset, limit int) ([]user.UserGroupMember, int64, error) {
	var members []user.UserGroupMember
	var total int64

	// Count total members
	if err := r.db.WithContext(ctx).Model(&user.UserGroupMember{}).Where("group_id = ?", groupID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count group members: %w", err)
	}

	// Get members with pagination
	err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Offset(offset).Limit(limit).Order("created_at DESC").Find(&members).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get group members with pagination: %w", err)
	}

	return members, total, nil
}

func (r *UserGroupRepository) GetGroupsByUserID(ctx context.Context, userID uint) ([]permission.UserGroup, error) {
	var groups []permission.UserGroup
	err := r.db.WithContext(ctx).
		Joins("JOIN user_group_members ON user_group_members.group_id = user_groups.id").
		Where("user_group_members.user_id = ?", userID).
		Find(&groups).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get groups by user id: %w", err)
	}
	return groups, nil
}

func (r *UserGroupRepository) IsPresetGroup(ctx context.Context, snowflakeID int64) (bool, error) {
	presetGroupIDs := []int64{
		1000000000000000001,
		1000000000000000002,
		1000000000000000003,
		1000000000000000004,
		1000000000000000005,
	}

	for _, id := range presetGroupIDs {
		if snowflakeID == id {
			return true, nil
		}
	}
	return false, nil
}
