package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrDuplicateUsername = errors.New("username already exists")
	ErrDuplicateEmail    = errors.New("email already exists")
)

type Repository interface {
	FindByID(ctx context.Context, id uint) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByIdentity(ctx context.Context, identity string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	UpdateLastLogin(ctx context.Context, id uint) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, page, pageSize int) ([]*User, int64, error)
}

type gormRepository struct { db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &gormRepository{db: db} }

func (r *gormRepository) FindByID(ctx context.Context, id uint) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrUserNotFound }
		return nil, fmt.Errorf("FindByID: %w", err)
	}
	return &u, nil
}

func (r *gormRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrUserNotFound }
		return nil, fmt.Errorf("FindByEmail: %w", err)
	}
	return &u, nil
}

func (r *gormRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrUserNotFound }
		return nil, fmt.Errorf("FindByUsername: %w", err)
	}
	return &u, nil
}

func (r *gormRepository) FindByIdentity(ctx context.Context, identity string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).Where("email = ? OR username = ?", identity, identity).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrUserNotFound }
		return nil, fmt.Errorf("FindByIdentity: %w", err)
	}
	return &u, nil
}

func (r *gormRepository) Create(ctx context.Context, user *User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "username") && isDuplicateKeyError(errStr) { return ErrDuplicateUsername }
		if strings.Contains(errStr, "email") && isDuplicateKeyError(errStr) { return ErrDuplicateEmail }
		return fmt.Errorf("Create user: %w", err)
	}
	return nil
}

func (r *gormRepository) Update(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *gormRepository) UpdateLastLogin(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("last_login_at", gorm.Expr("NOW()")).Error
}

func (r *gormRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}

func (r *gormRepository) List(ctx context.Context, page, pageSize int) ([]*User, int64, error) {
	var users []*User; var total int64
	query := r.db.WithContext(ctx).Model(&User{})
	if err := query.Count(&total).Error; err != nil { return nil, 0, err }
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func isDuplicateKeyError(errMsg string) bool {
	return strings.Contains(errMsg, "Duplicate entry") || strings.Contains(errMsg, "duplicate key")
}
