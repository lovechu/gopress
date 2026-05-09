package taxonomy

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var ErrTermNotFound = errors.New("term not found")
var ErrDuplicateSlug = errors.New("slug already exists")

type Repository interface {
	FindByID(ctx context.Context, id uint) (*Term, error)
	FindBySlug(ctx context.Context, slug string) (*Term, error)
	List(ctx context.Context, taxonomy string, page, pageSize int) ([]*Term, int64, error)
	Create(ctx context.Context, term *Term) error
	Update(ctx context.Context, term *Term) error
	Delete(ctx context.Context, id uint) error
	IncrementCount(ctx context.Context, id uint) error
	DecrementCount(ctx context.Context, id uint) error
}

type gormRepository struct { db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &gormRepository{db: db} }

func (r *gormRepository) FindByID(ctx context.Context, id uint) (*Term, error) {
	var t Term
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrTermNotFound }
		return nil, err
	}
	return &t, nil
}

func (r *gormRepository) FindBySlug(ctx context.Context, slug string) (*Term, error) {
	var t Term
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrTermNotFound }
		return nil, err
	}
	return &t, nil
}

func (r *gormRepository) List(ctx context.Context, taxonomy string, page, pageSize int) ([]*Term, int64, error) {
	var terms []*Term; var total int64
	query := r.db.WithContext(ctx).Model(&Term{})
	if taxonomy != "" { query = query.Where("taxonomy = ?", taxonomy) }
	if err := query.Count(&total).Error; err != nil { return nil, 0, err }
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("name ASC").Find(&terms).Error; err != nil {
		return nil, 0, err
	}
	return terms, total, nil
}

func (r *gormRepository) Create(ctx context.Context, term *Term) error {
	return r.db.WithContext(ctx).Create(term).Error
}

func (r *gormRepository) Update(ctx context.Context, term *Term) error {
	return r.db.WithContext(ctx).Save(term).Error
}

func (r *gormRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Term{}, id).Error
}

func (r *gormRepository) IncrementCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&Term{}).Where("id = ?", id).Update("count", gorm.Expr("count + 1")).Error
}

func (r *gormRepository) DecrementCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&Term{}).Where("id = ?", id).Update("count", gorm.Expr("GREATEST(count - 1, 0)")).Error
}
