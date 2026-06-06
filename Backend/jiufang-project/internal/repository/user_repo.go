package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"jiufang/internal/model/user"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetBySnowflakeID(ctx context.Context, snowflakeID int64) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("snowflake_id = ?", snowflakeID).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by snowflake_id: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) List(ctx context.Context, offset, limit int, username, role string, status int) ([]user.User, int64, error) {
	var users []user.User
	var total int64

	query := r.db.WithContext(ctx).Model(&user.User{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if status != -1 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	result := r.db.WithContext(ctx).Save(u)
	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&user.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uint, status int) error {
	result := r.db.WithContext(ctx).Model(&user.User{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update user status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uint, password string) error {
	result := r.db.WithContext(ctx).Model(&user.User{}).Where("id = ?", id).Update("password", password)
	if result.Error != nil {
		return fmt.Errorf("failed to update user password: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepository) UpdateAvatar(ctx context.Context, id uint, avatar string) error {
	result := r.db.WithContext(ctx).Model(&user.User{}).Where("id = ?", id).Update("avatar", avatar)
	if result.Error != nil {
		return fmt.Errorf("failed to update user avatar: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepository) UpdateFirstLoginStatus(ctx context.Context, id uint, isFirstLogin bool) error {
	result := r.db.WithContext(ctx).Model(&user.User{}).Where("id = ?", id).Update("is_first_login", isFirstLogin)
	if result.Error != nil {
		return fmt.Errorf("failed to update user first login status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}
