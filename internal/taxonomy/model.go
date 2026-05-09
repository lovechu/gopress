package taxonomy

import "gorm.io/gorm"

// Term 分类法术语（分类或标签）
type Term struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"type:varchar(128);not null" json:"name"`
	Slug        string         `gorm:"type:varchar(128);uniqueIndex;not null" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	Taxonomy    string         `gorm:"type:varchar(32);not null;index:idx_taxonomy_slug" json:"taxonomy"` // "category" or "tag"
	ParentID    *uint          `gorm:"index" json:"parent_id,omitempty"`
	Count       int            `gorm:"not null;default:0" json:"count"`
	CreatedAt   string         `gorm:"not null" json:"created_at"`
	UpdatedAt   string         `gorm:"not null" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Term) TableName() string { return "terms" }

type CreateTermRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	Slug        string `json:"slug" binding:"required,min=1,max=128"`
	Description string `json:"description"`
	Taxonomy    string `json:"taxonomy" binding:"required,oneof=category tag"`
	ParentID    *uint  `json:"parent_id"`
}

type UpdateTermRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=128"`
	Slug        *string `json:"slug" binding:"omitempty,min=1,max=128"`
	Description *string `json:"description"`
	ParentID    *uint   `json:"parent_id"`
}

type TermDTO struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Taxonomy    string `json:"taxonomy"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	Count       int    `json:"count"`
}

func (t *Term) ToDTO() TermDTO {
	return TermDTO{ID: t.ID, Name: t.Name, Slug: t.Slug, Description: t.Description,
		Taxonomy: t.Taxonomy, ParentID: t.ParentID, Count: t.Count}
}
