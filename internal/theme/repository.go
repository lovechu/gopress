package theme

import (
	"context"

	"gorm.io/gorm"
)

// gormRepository GORM主题仓库实现
type gormRepository struct {
	db *gorm.DB
}

// NewRepository 创建主题仓库
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// FindBySlug 根据slug查找主题
func (r *gormRepository) FindBySlug(ctx context.Context, slug string) (*Theme, error) {
	var theme Theme
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&theme).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrThemeNotFound
		}
		return nil, err
	}
	return &theme, nil
}

// FindActive 获取激活主题
func (r *gormRepository) FindActive(ctx context.Context) (*Theme, error) {
	var theme Theme
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).First(&theme).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrThemeNotFound
		}
		return nil, err
	}
	return &theme, nil
}

// List 列出所有主题
func (r *gormRepository) List(ctx context.Context) ([]*Theme, error) {
	var themes []*Theme
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&themes).Error; err != nil {
		return nil, err
	}
	return themes, nil
}

// SetActive 设置激活主题
func (r *gormRepository) SetActive(ctx context.Context, id uint) error {
	// 取消所有主题的激活状态
	if err := r.db.WithContext(ctx).Model(&Theme{}).Update("is_active", false).Error; err != nil {
		return err
	}
	// 设置指定主题为激活状态
	return r.db.WithContext(ctx).Model(&Theme{}).Where("id = ?", id).Update("is_active", true).Error
}

// Create 创建主题记录
func (r *gormRepository) Create(ctx context.Context, theme *Theme) error {
	return r.db.WithContext(ctx).Create(theme).Error
}

// Delete 删除主题
func (r *gormRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Theme{}, id).Error
}
