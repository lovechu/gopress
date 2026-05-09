package content

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var ErrPostNotFound = errors.New("post not found")

type Repository interface {
	FindByID(ctx context.Context, id uint, preload ...string) (*Post, error)
	FindBySlug(ctx context.Context, slug string) (*Post, error)
	List(ctx context.Context, filter ListFilter) ([]*Post, int64, error)
	Create(ctx context.Context, post *Post) error
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id uint) error
	AttachTerms(ctx context.Context, postID uint, termIDs []uint) error
	DetachAllTerms(ctx context.Context, postID uint) error
	CreateRevision(ctx context.Context, rev *PostRevision) error
	ListRevisions(ctx context.Context, postID uint) ([]*PostRevision, error)
}

type ListFilter struct {
	Status   string
	Type     string
	Slug     *string
	AuthorID uint
	TermIDs  []uint
	Page     int
	PageSize int
}

type gormRepository struct { db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &gormRepository{db: db} }

func (r *gormRepository) FindByID(ctx context.Context, id uint, preload ...string) (*Post, error) {
	var p Post
	q := r.db.WithContext(ctx)
	for _, p := range preload { q = q.Preload(p) }
	if err := q.First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrPostNotFound }
		return nil, err
	}
	return &p, nil
}

func (r *gormRepository) FindBySlug(ctx context.Context, slug string) (*Post, error) {
	var p Post
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrPostNotFound }
		return nil, err
	}
	return &p, nil
}

func (r *gormRepository) List(ctx context.Context, filter ListFilter) ([]*Post, int64, error) {
	var posts []*Post; var total int64
	q := r.db.WithContext(ctx).Model(&Post{}).Distinct("posts.*")
	if filter.Status != "" { q = q.Where("status = ?", filter.Status) }
	if filter.Type != "" { q = q.Where("type = ?", filter.Type) }
	if filter.Slug != nil { q = q.Where("slug = ?", *filter.Slug) }
	if filter.AuthorID > 0 { q = q.Where("author_id = ?", filter.AuthorID) }
	if len(filter.TermIDs) > 0 {
		q = q.Joins("JOIN post_terms ON post_terms.post_id = posts.id").
			Where("post_terms.term_id IN ?", filter.TermIDs)
	}
	if err := q.Count(&total).Error; err != nil { return nil, 0, err }
	offset := (filter.Page - 1) * filter.PageSize
	if err := q.Offset(offset).Limit(filter.PageSize).Order("created_at DESC").Find(&posts).Error; err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

func (r *gormRepository) Create(ctx context.Context, post *Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *gormRepository) Update(ctx context.Context, post *Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *gormRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Post{}, id).Error
}

func (r *gormRepository) AttachTerms(ctx context.Context, postID uint, termIDs []uint) error {
	var terms []PostTerm
	for _, tid := range termIDs {
		terms = append(terms, PostTerm{PostID: postID, TermID: tid})
	}
	return r.db.WithContext(ctx).Create(&terms).Error
}

func (r *gormRepository) DetachAllTerms(ctx context.Context, postID uint) error {
	return r.db.WithContext(ctx).Where("post_id = ?", postID).Delete(&PostTerm{}).Error
}

func (r *gormRepository) CreateRevision(ctx context.Context, rev *PostRevision) error {
	return r.db.WithContext(ctx).Create(rev).Error
}

func (r *gormRepository) ListRevisions(ctx context.Context, postID uint) ([]*PostRevision, error) {
	var revs []*PostRevision
	if err := r.db.WithContext(ctx).Where("post_id = ?", postID).
		Order("revision_number DESC").Find(&revs).Error; err != nil {
		return nil, err
	}
	return revs, nil
}
