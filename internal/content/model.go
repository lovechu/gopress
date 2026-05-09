package content

import (
	"time"

	"github.com/yourorg/gopress/internal/taxonomy"
	"gorm.io/gorm"
)

// ContentStatus 内容状态
type ContentStatus string

const (
	StatusDraft     ContentStatus = "draft"
	StatusPublished ContentStatus = "published"
	StatusTrash     ContentStatus = "trash"
)

// ContentType 内容类型
type ContentType string

const (
	TypePost ContentType = "post"
	TypePage ContentType = "page"
)

// Post 文章/页面模型
type Post struct {
	ID             uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Title          string         `gorm:"type:varchar(255);not null" json:"title"`
	Slug           string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Content        string         `gorm:"type:longtext" json:"content"`
	Excerpt        string         `gorm:"type:text" json:"excerpt"`
	Status         ContentStatus  `gorm:"type:varchar(32);not null;default:'draft'" json:"status"`
	Type           ContentType    `gorm:"type:varchar(32);not null;default:'post';index" json:"type"`
	AuthorID       uint           `gorm:"not null;index" json:"author_id"`
	CommentAllowed bool           `gorm:"not null;default:true" json:"comment_allowed"`
	ViewCount      int            `gorm:"not null;default:0" json:"view_count"`
	PublishedAt    *time.Time     `gorm:"index" json:"published_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	// 多对多：分类和标签
	Terms  []taxonomy.Term `gorm:"many2many:post_terms;" json:"terms,omitempty"`
}

func (Post) TableName() string { return "posts" }

// PostTerm 文章-术语关联表（含排序）
type PostTerm struct {
	PostID uint `gorm:"primaryKey"`
	TermID uint `gorm:"primaryKey"`
	Sort   int  `gorm:"not null;default:0"`
}

func (PostTerm) TableName() string { return "post_terms" }

// PostRevision 文章版本历史
type PostRevision struct {
	ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID    uint       `gorm:"not null;index" json:"post_id"`
	Title     string     `gorm:"type:varchar(255)" json:"title"`
	Content   string     `gorm:"type:longtext" json:"content"`
	Excerpt   string     `gorm:"type:text" json:"excerpt"`
	Status    ContentStatus `gorm:"type:varchar(32)" json:"status"`
	RevisionNumber int `gorm:"not null" json:"revision_number"`
	ChangedBy uint       `gorm:"not null" json:"changed_by"`
	CreatedAt time.Time  `json:"created_at"`
}

func (PostRevision) TableName() string { return "post_revisions" }

// ---- DTO ----

type CreatePostRequest struct {
	Title    string   `json:"title" binding:"required,min=1,max=255"`
	Slug     string   `json:"slug" binding:"required,min=1,max=255"`
	Content  string   `json:"content" binding:"required"`
	Excerpt  string   `json:"excerpt"`
	Status   string   `json:"status" binding:"omitempty,oneof=draft published"`
	Type     string   `json:"type" binding:"omitempty,oneof=post page"`
	TermIDs  []uint   `json:"term_ids"`
}

type UpdatePostRequest struct {
	Title    *string  `json:"title" binding:"omitempty,min=1,max=255"`
	Slug     *string  `json:"slug" binding:"omitempty,min=1,max=255"`
	Content  *string  `json:"content"`
	Excerpt  *string  `json:"excerpt"`
	Status   *string  `json:"status" binding:"omitempty,oneof=draft published"`
	TermIDs  []uint   `json:"term_ids"`
}

type PostDTO struct {
	ID             uint                `json:"id"`
	Title          string              `json:"title"`
	Slug           string              `json:"slug"`
	Content        string              `json:"content,omitempty"`
	Excerpt        string              `json:"excerpt"`
	Status         ContentStatus       `json:"status"`
	Type           ContentType         `json:"type"`
	AuthorID       uint                `json:"author_id"`
	AuthorName     string              `json:"author_name,omitempty"`
	CommentAllowed bool                `json:"comment_allowed"`
	ViewCount      int                 `json:"view_count"`
	Terms          []taxonomy.TermDTO  `json:"terms,omitempty"`
	PublishedAt    *time.Time          `json:"published_at,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

func (p *Post) ToDTO(authorName string) PostDTO {
	return PostDTO{
		ID: p.ID, Title: p.Title, Slug: p.Slug, Content: p.Content,
		Excerpt: p.Excerpt, Status: p.Status, Type: p.Type,
		AuthorID: p.AuthorID, AuthorName: authorName,
		CommentAllowed: p.CommentAllowed, ViewCount: p.ViewCount,
		PublishedAt: p.PublishedAt, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}
