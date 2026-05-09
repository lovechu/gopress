package user

import (
	"time"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin      Role = "admin"
	RoleEditor     Role = "editor"
	RoleAuthor     Role = "author"
	RoleSubscriber Role = "subscriber"
)

var RoleHierarchy = map[Role]int{
	RoleSubscriber: 1,
	RoleAuthor:     2,
	RoleEditor:     3,
	RoleAdmin:      4,
}

func (r Role) HasPermission(required Role) bool {
	return RoleHierarchy[r] >= RoleHierarchy[required]
}

var AllRoles = map[Role]bool{
	RoleAdmin: true, RoleEditor: true,
	RoleAuthor: true, RoleSubscriber: true,
}

type User struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Email        string         `gorm:"type:varchar(128);uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"type:varchar(256);not null" json:"-"`
	DisplayName  string         `gorm:"type:varchar(128);not null;default:''" json:"display_name"`
	Avatar       string         `gorm:"type:varchar(512);not null;default:''" json:"avatar"`
	Role         Role           `gorm:"type:varchar(32);not null;default:'subscriber'" json:"role"`
	Bio          string         `gorm:"type:text" json:"bio"`
	IsActive     bool           `gorm:"not null;default:true" json:"is_active"`
	LastLoginAt  *time.Time     `gorm:"index" json:"last_login_at"`
	CreatedAt    time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string { return "users" }

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Identity string `json:"identity" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" binding:"omitempty,max=128"`
	Bio         string `json:"bio"          binding:"omitempty,max=500"`
	Avatar      string `json:"avatar"       binding:"omitempty,url"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}

type UserDTO struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	Avatar      string     `json:"avatar"`
	Role        Role       `json:"role"`
	Bio         string     `json:"bio"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (u *User) ToDTO() *UserDTO {
	return &UserDTO{
		ID: u.ID, Username: u.Username, Email: u.Email,
		DisplayName: u.DisplayName, Avatar: u.Avatar,
		Role: u.Role, Bio: u.Bio, IsActive: u.IsActive,
		LastLoginAt: u.LastLoginAt, CreatedAt: u.CreatedAt,
	}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}
