import os

base = r"C:\Users\ichuy\WorkBuddy\20260505214615\gopress"

files = {
    # ========== user 模块补齐 ==========
    "internal/user/repository.go": r"""package user

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
""",

    "internal/user/service.go": r"""package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/yourorg/gopress/pkg/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is inactive")
	ErrWrongPassword      = errors.New("wrong password")
)

type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*UserDTO, error)
	Login(ctx context.Context, req LoginRequest) (*TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	GetProfile(ctx context.Context, userID uint) (*UserDTO, error)
	UpdateProfile(ctx context.Context, userID uint, req UpdateProfileRequest) (*UserDTO, error)
	ChangePassword(ctx context.Context, userID uint, req ChangePasswordRequest) error
	List(ctx context.Context, page, pageSize int) ([]*UserDTO, int64, error)
}

type service struct { repo Repository; jwtCfg jwt.Config; log *zap.Logger }

func NewService(repo Repository, jwtCfg jwt.Config, log *zap.Logger) Service {
	return &service{repo: repo, jwtCfg: jwtCfg, log: log}
}

func (s *service) Register(ctx context.Context, req RegisterRequest) (*UserDTO, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil { return nil, fmt.Errorf("hash password: %w", err) }
	u := &User{
		Username: req.Username, Email: req.Email, PasswordHash: string(hash),
		DisplayName: req.Username, Role: RoleSubscriber, IsActive: true,
	}
	if err := s.repo.Create(ctx, u); err != nil { return nil, err }
	s.log.Info("user registered", zap.Uint("id", u.ID))
	return u.ToDTO(), nil
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*TokenPair, error) {
	u, err := s.repo.FindByIdentity(ctx, req.Identity)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) { return nil, ErrInvalidCredentials }
		return nil, err
	}
	if !u.IsActive { return nil, ErrUserInactive }
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	go func() { _ = s.repo.UpdateLastLogin(context.Background(), u.ID) }()
	return s.generateTokenPair(u)
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := jwt.ParseRefreshToken(refreshToken, s.jwtCfg.RefreshSecret)
	if err != nil { return nil, fmt.Errorf("invalid refresh token: %w", err) }
	u, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil { return nil, err }
	if !u.IsActive { return nil, ErrUserInactive }
	return s.generateTokenPair(u)
}

func (s *service) GetProfile(ctx context.Context, userID uint) (*UserDTO, error) {
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil { return nil, err }
	return u.ToDTO(), nil
}

func (s *service) UpdateProfile(ctx context.Context, userID uint, req UpdateProfileRequest) (*UserDTO, error) {
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil { return nil, err }
	if req.DisplayName != "" { u.DisplayName = req.DisplayName }
	if req.Bio != "" { u.Bio = req.Bio }
	if req.Avatar != "" { u.Avatar = req.Avatar }
	if err := s.repo.Update(ctx, u); err != nil { return nil, err }
	return u.ToDTO(), nil
}

func (s *service) ChangePassword(ctx context.Context, userID uint, req ChangePasswordRequest) error {
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil { return err }
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.OldPassword)); err != nil {
		return ErrWrongPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil { return fmt.Errorf("hash password: %w", err) }
	u.PasswordHash = string(hash)
	return s.repo.Update(ctx, u)
}

func (s *service) List(ctx context.Context, page, pageSize int) ([]*UserDTO, int64, error) {
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 100 { pageSize = 20 }
	users, total, err := s.repo.List(ctx, page, pageSize)
	if err != nil { return nil, 0, err }
	dtos := make([]*UserDTO, len(users))
	for i, u := range users { dtos[i] = u.ToDTO() }
	return dtos, total, nil
}

func (s *service) generateTokenPair(u *User) (*TokenPair, error) {
	access, err := jwt.GenerateAccessToken(s.jwtCfg, u.ID, u.Username, string(u.Role))
	if err != nil { return nil, fmt.Errorf("generate access token: %w", err) }
	refresh, err := jwt.GenerateRefreshToken(s.jwtCfg, u.ID, u.Username, string(u.Role))
	if err != nil { return nil, fmt.Errorf("generate refresh token: %w", err) }
	return &TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: int64(s.jwtCfg.AccessExpire)}, nil
}
""",

    "internal/user/handler.go": r"""package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/gopress/pkg/response"
)

type Handler struct { svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.POST("/auth/register", h.Register)
	rg.POST("/auth/login", h.Login)
	rg.POST("/auth/refresh", h.RefreshToken)
}

func (h *Handler) RegisterAuthRoutes(rg *gin.RouterGroup) {
	rg.GET("/users/me", h.GetProfile)
	rg.PUT("/users/me", h.UpdateProfile)
	rg.POST("/users/me/password", h.ChangePassword)
}

func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/users", h.ListUsers)
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Fail(c, http.StatusBadRequest, err.Error()); return }
	dto, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrDuplicateUsername) { response.Fail(c, http.StatusConflict, "username already exists") }
		else if errors.Is(err, ErrDuplicateEmail) { response.Fail(c, http.StatusConflict, "email already exists") }
		else { response.Fail(c, http.StatusInternalServerError, "registration failed") }
		return
	}
	response.Created(c, dto)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Fail(c, http.StatusBadRequest, err.Error()); return }
	pair, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) { response.Fail(c, http.StatusUnauthorized, "invalid credentials") }
		else if errors.Is(err, ErrUserInactive) { response.Fail(c, http.StatusForbidden, "account is disabled") }
		else { response.Fail(c, http.StatusInternalServerError, "login failed") }
		return
	}
	response.OK(c, pair)
}

func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Fail(c, http.StatusBadRequest, err.Error()); return }
	pair, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil { response.Fail(c, http.StatusUnauthorized, "invalid or expired refresh token"); return }
	response.OK(c, pair)
}

func (h *Handler) GetProfile(c *gin.Context) {
	u, _ := c.Get("user_id")
	userID := u.(uint)
	dto, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) { response.Fail(c, http.StatusNotFound, "user not found") }
		else { response.Fail(c, http.StatusInternalServerError, "failed to get profile") }
		return
	}
	response.OK(c, dto)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Fail(c, http.StatusBadRequest, err.Error()); return }
	u, _ := c.Get("user_id")
	userID := u.(uint)
	dto, err := h.svc.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil { response.Fail(c, http.StatusInternalServerError, "update failed"); return }
	response.OK(c, dto)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Fail(c, http.StatusBadRequest, err.Error()); return }
	u, _ := c.Get("user_id")
	userID := u.(uint)
	if err := h.svc.ChangePassword(c.Request.Context(), userID, req); err != nil {
		if errors.Is(err, ErrWrongPassword) { response.Fail(c, http.StatusUnauthorized, "wrong password") }
		else { response.Fail(c, http.StatusInternalServerError, "change password failed") }
		return
	}
	response.OK(c, gin.H{"message": "password changed"})
}

func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	dtos, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil { response.Fail(c, http.StatusInternalServerError, "failed to list users"); return }
	response.Page(c, dtos, total, page, pageSize)
}

func mustGetUserID(c *gin.Context) uint {
	u, _ := c.Get("user_id")
	return u.(uint)
}
""",

    # ========== taxonomy 模块 ==========
    "internal/taxonomy/model.go": r"""package taxonomy

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
""",

    "internal/taxonomy/repository.go": r"""package taxonomy

import (
	"context"
	"errors"
	"fmt"

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
""",

    "internal/taxonomy/service.go": r"""package taxonomy

import (
	"context"
	"errors"
	"fmt"
)

var ErrInvalidTaxonomy = errors.New("invalid taxonomy type")

type Service interface {
	Create(ctx context.Context, req CreateTermRequest) (*TermDTO, error)
	Update(ctx context.Context, id uint, req UpdateTermRequest) (*TermDTO, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*TermDTO, error)
	List(ctx context.Context, taxonomy string, page, pageSize int) ([]*TermDTO, int64, error)
}

type service struct { repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) Create(ctx context.Context, req CreateTermRequest) (*TermDTO, error) {
	if req.Taxonomy != "category" && req.Taxonomy != "tag" {
		return nil, ErrInvalidTaxonomy
	}
	term := &Term{Name: req.Name, Slug: req.Slug, Description: req.Description,
		Taxonomy: req.Taxonomy, ParentID: req.ParentID}
	if err := s.repo.Create(ctx, term); err != nil { return nil, fmt.Errorf("create term: %w", err) }
	return term.ToDTO(), nil
}

func (s *service) Update(ctx context.Context, id uint, req UpdateTermRequest) (*TermDTO, error) {
	term, err := s.repo.FindByID(ctx, id)
	if err != nil { return nil, err }
	if req.Name != nil { term.Name = *req.Name }
	if req.Slug != nil { term.Slug = *req.Slug }
	if req.Description != nil { term.Description = *req.Description }
	if req.ParentID != nil { term.ParentID = req.ParentID }
	if err := s.repo.Update(ctx, term); err != nil { return nil, err }
	return term.ToDTO(), nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) GetByID(ctx context.Context, id uint) (*TermDTO, error) {
	term, err := s.repo.FindByID(ctx, id)
	if err != nil { return nil, err }
	return term.ToDTO(), nil
}

func (s *service) List(ctx context.Context, taxonomy string, page, pageSize int) ([]*TermDTO, int64, error) {
	terms, total, err := s.repo.List(ctx, taxonomy, page, pageSize)
	if err != nil { return nil, 0, err }
	dtos := make([]*TermDTO, len(terms))
	for i, t := range terms { dtos[i] = t.ToDTO() }
	return dtos, total, nil
}
""",

    "internal/taxonomy/handler.go": r"""package taxonomy

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/gopress/pkg/response"
)

type Handler struct { svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/terms", h.ListTerms)
	rg.GET("/terms/:id", h.GetTerm)
}

func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.POST("/terms", h.CreateTerm)
	rg.PUT("/terms/:id", h.UpdateTerm)
	rg.DELETE("/terms/:id", h.DeleteTerm)
}

func (h *Handler) CreateTerm(c *gin.Context) {
	var req CreateTermRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Fail(c, 400, err.Error()); return }
	dto, err := h.svc.Create(c.Request.Context(), req)
	if err != nil { response.Fail(c, 500, err.Error()); return }
	response.Created(c, dto)
}

func (h *Handler) UpdateTerm(c *gin.Context) {
	id, _ := strconv.ParsUint(c.Param("id"), 10, 32)
	var req UpdateTermRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Fail(c, 400, err.Error()); return }
	dto, err := h.svc.Update(c.Request.Context(), uint(id), req)
	if err != nil { response.Fail(c, 500, err.Error()); return }
	response.OK(c, dto)
}

func (h *Handler) DeleteTerm(c *gin.Context) {
	id, _ := strconv.ParsUint(c.Param("id"), 10, 32)
	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil { response.Fail(c, 500, err.Error()); return }
	response.OK(c, gin.H{"message": "deleted"})
}

func (h *Handler) GetTerm(c *gin.Context) {
	id, _ := strconv.ParsUint(c.Param("id"), 10, 32)
	dto, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil { response.Fail(c, 404, "term not found"); return }
	response.OK(c, dto)
}

func (h *Handler) ListTerms(c *gin.Context) {
	taxonomy := c.Query("taxonomy")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	dtos, total, err := h.svc.List(c.Request.Context(), taxonomy, page, pageSize)
	if err != nil { response.Fail(c, 500, err.Error()); return }
	response.Page(c, dtos, total, page, pageSize)
}
""",

    # ========== content 模块 ==========
    "internal/content/model.go": r"""package content

import (
	"time"

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
""",

    "internal/content/repository.go": r"""package content

import (
	"context"
	"errors"
	"fmt"

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
""",

    "internal/content/service.go": r"""package content

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yourorg/gopress/internal/taxonomy"
	"github.com/yourorg/gopress/internal/user"
	"gorm.io/gorm"
)

var (
	ErrForbidden      = errors.New("forbidden: not the author")
	ErrInvalidStatus  = errors.New("invalid status")
)

type Service interface {
	Create(ctx context.Context, authorID uint, req CreatePostRequest) (*PostDTO, error)
	Update(ctx context.Context, postID, actorID uint, actorRole user.Role, req UpdatePostRequest) (*PostDTO, error)
	Delete(ctx context.Context, postID, actorID uint, actorRole user.Role) error
	GetByID(ctx context.Context, id uint) (*PostDTO, error)
	List(ctx context.Context, filter ListFilter) ([]*PostDTO, int64, error)
	ListRevisions(ctx context.Context, postID uint) ([]*PostRevision, error)
}

type service struct {
	repo      Repository
	termRepo  taxonomy.Repository
	userRepo  user.Repository
}

func NewService(repo Repository, termRepo taxonomy.Repository, userRepo user.Repository) Service {
	return &service{repo: repo, termRepo: termRepo, userRepo: userRepo}
}

func (s *service) Create(ctx context.Context, authorID uint, req CreatePostRequest) (*PostDTO, error) {
	status := ContentStatus(req.Status)
	if req.Status == "" { status = StatusDraft }
	if status != StatusDraft && status != StatusPublished { return nil, ErrInvalidStatus }

	post := &Post{
		Title: req.Title, Slug: req.Slug, Content: req.Content,
		Excerpt: req.Excerpt, Status: status, Type: TypePost,
		AuthorID: authorID,
	}
	if req.Type == "page" { post.Type = TypePage }
	if status == StatusPublished {
		now := time.Now()
		post.PublishedAt = &now
	}

	if err := s.repo.Create(ctx, post); err != nil { return nil, fmt.Errorf("create post: %w", err) }

	// 保存修订版本
	s.saveRevision(ctx, post, authorID)

	// 关联术语
	if len(req.TermIDs) > 0 {
		_ = s.repo.DetachAllTerms(ctx, post.ID)
		_ = s.repo.AttachTerms(ctx, post.ID, req.TermIDs)
	}
	return s.loadPostDTO(ctx, post)
}

func (s *service) Update(ctx context.Context, postID, actorID uint, actorRole user.Role, req UpdatePostRequest) (*PostDTO, error) {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil { return nil, err }

	// 权限检查：作者只能改自己的文章，editor/admin 可以改任何人的
	if actorRole != user.RoleAdmin && actorRole != user.RoleEditor {
		if post.AuthorID != actorID { return nil, ErrForbidden }
	}

	if req.Title != nil { post.Title = *req.Title }
	if req.Slug != nil { post.Slug = *req.Slug }
	if req.Content != nil { post.Content = *req.Content }
	if req.Excerpt != nil { post.Excerpt = *req.Excerpt }
	if req.Status != nil {
		newStatus := ContentStatus(*req.Status)
		if newStatus == StatusPublished && post.PublishedAt == nil {
			now := time.Now()
			post.PublishedAt = &now
		}
		post.Status = newStatus
	}

	if err := s.repo.Update(ctx, post); err != nil { return nil, err }

	s.saveRevision(ctx, post, actorID)

	if req.TermIDs != nil {
		_ = s.repo.DetachAllTerms(ctx, post.ID)
		if len(req.TermIDs) > 0 {
			_ = s.repo.AttachTerms(ctx, post.ID, req.TermIDs)
		}
	}
	return s.loadPostDTO(ctx, post)
}

func (s *service) Delete(ctx context.Context, postID, actorID uint, actorRole user.Role) error {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil { return err }
	if actorRole != user.RoleAdmin && actorRole != user.RoleEditor {
		if post.AuthorID != actorID { return ErrForbidden }
	}
	return s.repo.Delete(ctx, postID)
}

func (s *service) GetByID(ctx context.Context, id uint) (*PostDTO, error) {
	post, err := s.repo.FindByID(ctx, id, "Terms")
	if err != nil { return nil, err }
	return s.loadPostDTO(ctx, post)
}

func (s *service) List(ctx context.Context, filter ListFilter) ([]*PostDTO, int64, error) {
	posts, total, err := s.repo.List(ctx, filter)
	if err != nil { return nil, 0, err }
	dtos := make([]*PostDTO, len(posts))
	for i, p := range posts {
		dto, _ := s.loadPostDTO(ctx, p)
		dtos[i] = dto
	}
	return dtos, total, nil
}

func (s *service) ListRevisions(ctx context.Context, postID uint) ([]*PostRevision, error) {
	return s.repo.ListRevisions(ctx, postID)
}

func (s *service) saveRevision(ctx context.Context, post *Post, changedBy uint) {
	// 获取当前最大版本号
	var rev PostRevision
	s.repo.(*gormRepository).db.WithContext(ctx).Where("post_id = ?", post.ID).
		Order("revision_number DESC").First(&rev)
	nextNum := rev.RevisionNumber + 1
	newRev := &PostRevision{
		PostID: post.ID, Title: post.Title, Content: post.Content,
		Excerpt: post.Excerpt, Status: post.Status,
		RevisionNumber: nextNum, ChangedBy: changedBy,
	}
	_ = s.repo.CreateRevision(ctx, newRev)
}

func (s *service) loadPostDTO(ctx context.Context, post *Post) (*PostDTO, error) {
	author, _ := s.userRepo.FindByID(ctx, post.AuthorID)
	authorName := ""
	if author != nil { authorName = author.DisplayName }
	// 手动加载 Terms
	var terms []taxonomy.Term
	s.repo.(*gormRepository).db.WithContext(ctx).
		Joins("JOIN post_terms ON post_terms.term_id = terms.id").
		Where("post_terms.post_id = ?", post.ID).Find(&terms)
	termDTOs := make([]taxonomy.TermDTO, len(terms))
	for i, t := range terms { termDTOs[i] = t.ToDTO() }
	dto := post.ToDTO(authorName)
	dto.Terms = termDTOs
	return dto, nil
}
""",

    "internal/content/handler.go": r"""package content

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/gopress/internal/user"
	"github.com/yourorg/gopress/pkg/response"
)

type Handler struct { svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/posts", h.ListPosts)
	rg.GET("/posts/:id", h.GetPost)
	rg.GET("/pages", h.ListPages)
	rg.GET("/pages/:id", h.GetPage)
}

func (h *Handler) RegisterAuthRoutes(rg *gin.RouterGroup) {
	rg.POST("/posts", h.CreatePost)
	rg.PUT("/posts/:id", h.UpdatePost)
	rg.DELETE("/posts/:id", h.DeletePost)
	rg.GET("/posts/:id/revisions", h.ListRevisions)
}

func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	// admin 权限已通过 middleware 完成
}

func getActor(c *gin.Context) (uint, user.Role) {
	u, _ := c.Get("user_id")
	r, _ := c.Get("user_role")
	return u.(uint), user.Role(r.(string))
}

func (h *Handler) CreatePost(c *gin.Context) {
	authorID, _ := getActor(c)
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Fail(c, 400, err.Error()); return }
	dto, err := h.svc.Create(c.Request.Context(), authorID, req)
	if err != nil { response.Fail(c, 500, err.Error()); return }
	response.Created(c, dto)
}

func (h *Handler) UpdatePost(c *gin.Context) {
	postID, _ := strconv.ParsUint(c.Param("id"), 10, 32)
	actorID, actorRole := getActor(c)
	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Fail(c, 400, err.Error()); return }
	dto, err := h.svc.Update(c.Request.Context(), uint(postID), actorID, actorRole, req)
	if err != nil {
		if err == ErrForbidden { response.Fail(c, 403, "forbidden") }
		else { response.Fail(c, 500, err.Error()) }
		return
	}
	response.OK(c, dto)
}

func (h *Handler) DeletePost(c *gin.Context) {
	postID, _ := strconv.ParsUint(c.Param("id"), 10, 32)
	actorID, actorRole := getActor(c)
	if err := h.svc.Delete(c.Request.Context(), uint(postID), actorID, actorRole); err != nil {
		if err == ErrForbidden { response.Fail(c, 403, "forbidden") }
		else { response.Fail(c, 500, err.Error()) }
		return
	}
	response.OK(c, gin.H{"message": "deleted"})
}

func (h *Handler) GetPost(c *gin.Context) {
	id, _ := strconv.ParsUint(c.Param("id"), 10, 32)
	dto, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil { response.Fail(c, 404, "post not found"); return }
	response.OK(c, dto)
}

func (h *Handler) ListPosts(c *gin.Context) { h.listContent(c, "post") }

func (h *Handler) ListPages(c *gin.Context) { h.listContent(c, "page") }

func (h *Handler) listContent(c *gin.Context, contentType string) {
	filter := ListFilter{Type: contentType, Page: 1, PageSize: 20}
	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 { filter.Page = page }
	if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 { filter.PageSize = pageSize }
	if status := c.Query("status"); status != "" { filter.Status = status }
	// 已登录用户看到自己的文章，未登录只看到 published
	if actorID, exists := c.Get("user_id"); exists {
		u, _ := actorID.(uint)
		role, _ := c.Get("user_role")
		if role != user.RoleAdmin && role != user.RoleEditor {
			filter.AuthorID = u
		}
	} else {
		filter.Status = string(StatusPublished)
	}
	dtos, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil { response.Fail(c, 500, err.Error()); return }
	response.Page(c, dtos, total, filter.Page, filter.PageSize)
}

func (h *Handler) ListRevisions(c *gin.Context) {
	postID, _ := strconv.ParsUint(c.Param("id"), 10, 32)
	revs, err := h.svc.ListRevisions(c.Request.Context(), uint(postID))
	if err != nil { response.Fail(c, 500, err.Error()); return }
	response.OK(c, revs)
}
""",

    # ========== migrations ==========
    "migrations/002_create_terms.up.sql": """CREATE TABLE IF NOT EXISTS terms (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(128) NOT NULL,
  slug VARCHAR(128) NOT NULL UNIQUE,
  description TEXT,
  taxonomy VARCHAR(32) NOT NULL,
  parent_id BIGINT UNSIGNED NULL,
  count INT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  INDEX idx_taxonomy_slug (taxonomy, slug),
  INDEX idx_parent (parent_id),
  INDEX idx_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
""",
    "migrations/002_create_terms.down.sql": "DROP TABLE IF EXISTS terms;\n",

    "migrations/003_create_posts.up.sql": """CREATE TABLE IF NOT EXISTS posts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  slug VARCHAR(255) NOT NULL UNIQUE,
  content LONGTEXT,
  excerpt TEXT,
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  type VARCHAR(32) NOT NULL DEFAULT 'post',
  author_id BIGINT UNSIGNED NOT NULL,
  comment_allowed TINYINT(1) NOT NULL DEFAULT 1,
  view_count INT NOT NULL DEFAULT 0,
  published_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  INDEX idx_author (author_id),
  INDEX idx_status (status),
  INDEX idx_type (type),
  INDEX idx_published (published_at),
  INDEX idx_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS post_terms (
  post_id BIGINT UNSIGNED NOT NULL,
  term_id BIGINT UNSIGNED NOT NULL,
  sort INT NOT NULL DEFAULT 0,
  PRIMARY KEY (post_id, term_id),
  INDEX idx_post (post_id),
  INDEX idx_term (term_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS post_revisions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  post_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(255),
  content LONGTEXT,
  excerpt TEXT,
  status VARCHAR(32),
  revision_number INT NOT NULL,
  changed_by BIGINT UNSIGNED NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_post (post_id),
  INDEX idx_revision (post_id, revision_number)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
""",
    "migrations/003_create_posts.down.sql": """DROP TABLE IF EXISTS post_revisions;
DROP TABLE IF EXISTS post_terms;
DROP TABLE IF EXISTS posts;
""",

    # ========== media 模块 ==========
    "internal/media/model.go": r"""package media

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// MediaType 媒体类型
type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
	MediaTypeOther   MediaType = "other"
)

// Media 媒体模型
type Media struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID          string    `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	FileName      string    `gorm:"type:varchar(255);not null" json:"file_name"`
	OriginalName  string    `gorm:"type:varchar(500);not null" json:"original_name"`
	FileSize      int64     `gorm:"type:bigint;not null;default:0" json:"file_size"`
	MIMEType      string    `gorm:"type:varchar(100);not null" json:"mime_type"`
	MediaType     MediaType `gorm:"type:varchar(20);not null;default:'other'" json:"media_type"`
	Width         int       `gorm:"not null;default:0" json:"width"`
	Height        int       `gorm:"not null;default:0" json:"height"`
	Alt           string    `gorm:"type:varchar(255);default:''" json:"alt"`
	Caption       string    `gorm:"type:text" json:"caption"`
	StorageKey    string    `gorm:"type:varchar(500);not null" json:"storage_key"`
	ThumbnailKeys string    `gorm:"type:varchar(1000);default:''" json:"-"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Media) TableName() string { return "media" }

// ThumbnailInfo 缩略图信息
type ThumbnailInfo map[string]string

func (m *Media) GetThumbnails() ThumbnailInfo {
	if m.ThumbnailKeys == "" {
		return ThumbnailInfo{}
	}
	var info ThumbnailInfo
	json.Unmarshal([]byte(m.ThumbnailKeys), &info)
	return info
}

func (m *Media) SetThumbnails(info ThumbnailInfo) {
	if info == nil {
		m.ThumbnailKeys = ""
		return
	}
	data, _ := json.Marshal(info)
	m.ThumbnailKeys = string(data)
}

// DetectMediaType 根据 MIME 类型判断媒体类型
func DetectMediaType(mime string) MediaType {
	switch {
	case len(mime) >= 5 && mime[:5] == "image":
		return MediaTypeImage
	case len(mime) >= 5 && mime[:5] == "video":
		return MediaTypeVideo
	case len(mime) >= 5 && mime[:5] == "audio":
		return MediaTypeAudio
	case mime == "application/pdf" ||
		mime == "application/msword" ||
		mime == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" ||
		mime == "application/vnd.ms-excel" ||
		mime == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return MediaTypeDocument
	default:
		return MediaTypeOther
	}
}

// IsImage 判断是否为图片
func IsImage(mime string) bool {
	return len(mime) >= 5 && mime[:5] == "image"
}

// ---- Request/Response DTOs ----

// MediaFilter 媒体筛选条件
type MediaFilter struct {
	Type     string `form:"type"`
	UserID   *uint  `form:"-"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// Normalize 标准化分页参数
func (f *MediaFilter) Normalize() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 || f.PageSize > 100 {
		f.PageSize = 20
	}
}

// MediaDTO 媒体响应DTO
type MediaDTO struct {
	ID           uint         `json:"id"`
	UUID         string       `json:"uuid"`
	FileName     string       `json:"file_name"`
	OriginalName string       `json:"original_name"`
	FileSize     int64        `json:"file_size"`
	MIMEType     string       `json:"mime_type"`
	MediaType    MediaType    `json:"media_type"`
	Width        int          `json:"width"`
	Height       int          `json:"height"`
	Alt          string       `json:"alt"`
	Caption      string       `json:"caption"`
	URL          string       `json:"url"`
	Thumbnails   ThumbnailInfo `json:"thumbnails,omitempty"`
	UserID       uint         `json:"user_id"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// UpdateMediaRequest 更新媒体请求
type UpdateMediaRequest struct {
	Alt      *string `json:"alt"`
	Caption  *string `json:"caption"`
}

// GenerateThumbnailsRequest 生成缩略图请求
type GenerateThumbnailsRequest struct {
	Sizes []string `json:"sizes"` // e.g. ["small", "medium", "large"]
}

// ToDTO 转换为DTO
func (m *Media) ToDTO(baseURL string) MediaDTO {
	thumbnails := m.GetThumbnails()
	result := make(ThumbnailInfo)
	for name, key := range thumbnails {
		result[name] = baseURL + "/" + key
	}
	return MediaDTO{
		ID: m.ID, UUID: m.UUID, FileName: m.FileName, OriginalName: m.OriginalName,
		FileSize: m.FileSize, MIMEType: m.MIMEType, MediaType: m.MediaType,
		Width: m.Width, Height: m.Height, Alt: m.Alt, Caption: m.Caption,
		URL: baseURL + "/" + m.StorageKey, Thumbnails: result,
		UserID: m.UserID, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}
""",

    "internal/media/storage.go": r"""package media

import (
	"context"
	"io"
)

// Storage 存储接口
type Storage interface {
	// Upload 上传文件，返回存储key和公开访问URL
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error)
	// Delete 删除文件
	Delete(ctx context.Context, key string) error
	// Exists 检查文件是否存在
	Exists(ctx context.Context, key string) (bool, error)
	// GetURL 获取文件访问URL
	GetURL(key string) string
	// Download 下载文件
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	// GetBaseURL 获取基础URL
	GetBaseURL() string
}

// ThumbnailSize 缩略图尺寸
type ThumbnailSize struct {
	Name   string
	Width  int
	Height int
}

// DefaultThumbnailSizes 默认缩略图尺寸
var DefaultThumbnailSizes = []ThumbnailSize{
	{Name: "small", Width: 150, Height: 150},
	{Name: "medium", Width: 300, Height: 300},
	{Name: "large", Width: 800, Height: 800},
}
""",

    "internal/media/local.go": r"""package media

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// LocalStorage 本地文件系统存储
type LocalStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage 创建本地存储
func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	// 确保目录存在
	os.MkdirAll(basePath, 0755)
	return &LocalStorage{basePath: basePath, baseURL: baseURL}
}

func (s *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	fullPath := filepath.Join(s.basePath, key)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create directory: %w", err)
	}

	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	// 复制内容
	written, err := io.Copy(file, reader)
	if err != nil {
		os.Remove(fullPath)
		return "", fmt.Errorf("write file: %w", err)
	}

	if written != size {
		os.Remove(fullPath)
		return "", fmt.Errorf("size mismatch: expected %d, written %d", size, written)
	}

	return key, nil
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(s.basePath, key)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

func (s *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := filepath.Join(s.basePath, key)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *LocalStorage) GetURL(key string) string {
	return s.baseURL + "/" + key
}

func (s *LocalStorage) GetBaseURL() string {
	return s.baseURL
}

func (s *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, key)
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, err
	}
	return file, nil
}

// EnsureLocalStorage middleware: 提供静态文件服务
func EnsureLocalStorage(basePath, baseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求路径去除 baseURL 前缀
		prefix := baseURL
		if len(prefix) > 0 && prefix[0] == '/' {
			prefix = prefix[1:]
		}
		path := c.Request.URL.Path
		if len(path) > len(prefix) && path[:len(prefix)+1] == prefix+"/" {
			filePath := path[len(prefix)+1:]
			fullPath := filepath.Join(basePath, filePath)
			if _, err := os.Stat(fullPath); err == nil {
				http.ServeFile(c.Writer, c.Request, fullPath)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
""",

    "internal/media/minio.go": r"""package media

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	Region    string
	BaseURL   string
}

// MinIOStorage MinIO对象存储
type MinIOStorage struct {
	client   *minio.Client
	bucket   string
	region   string
	baseURL  string
}

// NewMinIOStorage 创建MinIO存储
func NewMinIOStorage(cfg MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio: %w", err)
	}

	// 确保bucket存在
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{Region: cfg.Region})
		if err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
	}

	return &MinIOStorage{
		client:  client,
		bucket:  cfg.Bucket,
		region:  cfg.Region,
		baseURL: cfg.BaseURL,
	}, nil
}

func (s *MinIOStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload to minio: %w", err)
	}
	return key, nil
}

func (s *MinIOStorage) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete from minio: %w", err)
	}
	return nil
}

func (s *MinIOStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *MinIOStorage) GetURL(key string) string {
	return s.baseURL + "/" + key
}

func (s *MinIOStorage) GetBaseURL() string {
	return s.baseURL
}

func (s *MinIOStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get from minio: %w", err)
	}
	return obj, nil
}
""",

    "internal/media/thumbnail.go": r"""package media

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"github.com/disintegration/imaging"
)

// ImageProcessor 图片处理器
type ImageProcessor struct{}

// NewImageProcessor 创建图片处理器
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
}

// GetImageDimensions 获取图片尺寸
func (p *ImageProcessor) GetImageDimensions(reader io.Reader) (int, int, error) {
	config, format, err := image.DecodeConfig(reader)
	if err != nil {
		return 0, 0, fmt.Errorf("decode config: %w", err)
	}
	_ = format // silence unused warning
	return config.Width, config.Height, nil
}

// GenerateThumbnail 生成单张缩略图
func (p *ImageProcessor) GenerateThumbnail(reader io.Reader, size ThumbnailSize) ([]byte, error) {
	// 解码图片
	img, err := imaging.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	// 缩放并裁剪到目标尺寸
	thumb := imaging.Thumbnail(img, size.Width, size.Height, imaging.Lanczos)

	// 编码为PNG
	var buf bytes.Buffer
	err = imaging.Encode(&buf, thumb, imaging.PNG)
	if err != nil {
		return nil, fmt.Errorf("encode thumbnail: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateThumbnails 生成多种尺寸缩略图
func (p *ImageProcessor) GenerateThumbnails(reader io.Reader, sizes []ThumbnailSize) (map[string][]byte, error) {
	results := make(map[string][]byte)

	for _, size := range sizes {
		// 需要重新读取，所以先把原始数据缓存
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("read data: %w", err)
		}

		thumb, err := p.GenerateThumbnail(bytes.NewReader(data), size)
		if err != nil {
			// 单个失败不影响其他
			continue
		}
		results[size.Name] = thumb
	}

	return results, nil
}

// GenerateAndSaveThumbnails 生成并保存缩略图
func (p *ImageProcessor) GenerateAndSaveThumbnails(ctx context.Context, storage Storage, sourceKey string, sizes []ThumbnailSize) (map[string]string, error) {
	// 下载原始文件
	reader, err := storage.Download(ctx, sourceKey)
	if err != nil {
		return nil, fmt.Errorf("download source: %w", err)
	}
	defer reader.Close()

	// 生成缩略图
	thumbs, err := p.GenerateThumbnails(reader, sizes)
	if err != nil {
		return nil, err
	}

	// 上传缩略图
	results := make(map[string]string)
	for name, data := range thumbs {
		// 构建缩略图key
		ext := ".png" // 缩略图统一用PNG
		dotIdx := strings.LastIndex(sourceKey, ".")
		if dotIdx > 0 {
			thumbKey := sourceKey[:dotIdx] + "_" + name + ext
			_, err := storage.Upload(ctx, thumbKey, bytes.NewReader(data), int64(len(data)), "image/png")
			if err != nil {
				continue
			}
			results[name] = thumbKey
		}
	}

	return results, nil
}

// DetectImageFormat 检测图片格式
func DetectImageFormat(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	switch {
	case data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return "jpeg"
	case data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47:
		return "png"
	case data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46:
		return "gif"
	case data[0] == 0x52 && data[1] == 0x49 && strings.HasPrefix(string(data[0:12]), "RIFF") && strings.HasSuffix(string(data[0:12]), "WEBP"):
		return "webp"
	}
	return ""
}

func init() {
	// 注册自定义图片格式解码（如果需要支持更多格式）
	_ = gif.GIF
	_ = jpeg.JPEG
	_ = png.PNG
}
""",

    "internal/media/repository.go": r"""package media

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var ErrMediaNotFound = errors.New("media not found")

type Repository interface {
	Create(ctx context.Context, media *Media) error
	FindByID(ctx context.Context, id uint) (*Media, error)
	FindByUUID(ctx context.Context, uuid string) (*Media, error)
	Update(ctx context.Context, media *Media) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter MediaFilter) ([]*Media, int64, error)
}

type gormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &gormRepository{db: db} }

func (r *gormRepository) Create(ctx context.Context, media *Media) error {
	return r.db.WithContext(ctx).Create(media).Error
}

func (r *gormRepository) FindByID(ctx context.Context, id uint) (*Media, error) {
	var m Media
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMediaNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *gormRepository) FindByUUID(ctx context.Context, uuid string) (*Media, error) {
	var m Media
	if err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMediaNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *gormRepository) Update(ctx context.Context, media *Media) error {
	return r.db.WithContext(ctx).Save(media).Error
}

func (r *gormRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Media{}, id).Error
}

func (r *gormRepository) List(ctx context.Context, filter MediaFilter) ([]*Media, int64, error) {
	var media []*Media
	var total int64

	query := r.db.WithContext(ctx).Model(&Media{})

	if filter.Type != "" {
		query = query.Where("media_type = ?", filter.Type)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Keyword != "" {
		keyword := "%" + filter.Keyword + "%"
		query = query.Where("original_name LIKE ? OR file_name LIKE ?", keyword, keyword)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	if err := query.Offset(offset).Limit(filter.PageSize).Order("created_at DESC").Find(&media).Error; err != nil {
		return nil, 0, err
	}

	return media, total, nil
}

// DeleteByStorageKey 根据存储key删除
func (r *gormRepository) DeleteByStorageKey(ctx context.Context, storageKey string) error {
	return r.db.WithContext(ctx).Where("storage_key = ?", storageKey).Delete(&Media{}).Error
}
""",

    "internal/media/service.go": r"""package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidFileType = errors.New("invalid file type")
	ErrFileTooLarge     = errors.New("file too large")
	ErrNotImage         = errors.New("not an image file")
	ErrForbidden        = errors.New("forbidden")
)

// Config 媒体模块配置
type Config struct {
	MaxFileSize   int64    // 最大文件大小（字节）
	AllowedTypes  []string // 允许的MIME类型
	ThumbnailDir  string   // 缩略图目录
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	MaxFileSize:  50 * 1024 * 1024, // 50MB
	AllowedTypes: []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	},
}

type Service interface {
	Upload(ctx context.Context, file *multipart.FileHeader, userID uint) (*Media, error)
	Delete(ctx context.Context, uuid string, userID uint, isAdmin bool) error
	GetByUUID(ctx context.Context, uuid string, userID uint, isAdmin bool) (*MediaDTO, error)
	List(ctx context.Context, filter MediaFilter) ([]*MediaDTO, int64, error)
	Update(ctx context.Context, uuid string, userID uint, isAdmin bool, req UpdateMediaRequest) (*MediaDTO, error)
	GenerateThumbnails(ctx context.Context, uuid string, userID uint, isAdmin bool, sizes []string) (*MediaDTO, error)
}

type service struct {
	repo      Repository
	storage   Storage
	config    Config
	processor *ImageProcessor
	baseURL   string
}

func NewService(repo Repository, storage Storage, cfg Config, baseURL string) Service {
	return &service{
		repo:      repo,
		storage:   storage,
		config:    cfg,
		processor: NewImageProcessor(),
		baseURL:   baseURL,
	}
}

func (s *service) Upload(ctx context.Context, file *multipart.FileHeader, userID uint) (*Media, error) {
	// 验证文件大小
	if file.Size > s.config.MaxFileSize {
		return nil, ErrFileTooLarge
	}

	// 验证文件类型
	contentType := file.Header.Get("Content-Type")
	if !s.isAllowedType(contentType) {
		return nil, ErrInvalidFileType
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	newUUID := uuid.New().String()
	fileName := newUUID + ext

	// 构建存储路径: media/{userID}/{uuid}.{ext}
	storageKey := fmt.Sprintf("media/%d/%s", userID, fileName)

	// 打开文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer src.Close()

	// 上传到存储
	_, err = s.storage.Upload(ctx, storageKey, src, file.Size, contentType)
	if err != nil {
		return nil, fmt.Errorf("upload: %w", err)
	}

	// 获取图片尺寸
	width, height := 0, 0
	if IsImage(contentType) {
		// 重新读取文件以获取尺寸
		src.Seek(0, io.SeekStart)
		data, _ := io.ReadAll(src)
		src.Seek(0, io.SeekStart)
		w, h, _ := s.processor.GetImageDimensions(strings.NewReader(string(data)))
		width, height = w, h
	}

	// 创建媒体记录
	media := &Media{
		UUID:         newUUID,
		FileName:     fileName,
		OriginalName: file.Filename,
		FileSize:     file.Size,
		MIMEType:     contentType,
		MediaType:    DetectMediaType(contentType),
		Width:        width,
		Height:       height,
		StorageKey:   storageKey,
		UserID:       userID,
	}

	if err := s.repo.Create(ctx, media); err != nil {
		// 回滚：删除已上传的文件
		s.storage.Delete(ctx, storageKey)
		return nil, fmt.Errorf("save media: %w", err)
	}

	return media, nil
}

func (s *service) Delete(ctx context.Context, uuid string, userID uint, isAdmin bool) error {
	media, err := s.repo.FindByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	// 权限检查：只有上传者或admin可以删除
	if !isAdmin && media.UserID != userID {
		return ErrForbidden
	}

	// 删除存储文件
	if err := s.storage.Delete(ctx, media.StorageKey); err != nil {
		// 记录错误但不阻塞删除
		fmt.Printf("delete storage file error: %v\n", err)
	}

	// 删除缩略图
	thumbs := media.GetThumbnails()
	for _, key := range thumbs {
		s.storage.Delete(ctx, key)
	}

	// 删除数据库记录
	return s.repo.Delete(ctx, media.ID)
}

func (s *service) GetByUUID(ctx context.Context, uuid string, userID uint, isAdmin bool) (*MediaDTO, error) {
	media, err := s.repo.FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	// 权限检查：非admin只能查看自己的
	if !isAdmin && media.UserID != userID {
		return nil, ErrForbidden
	}

	dto := media.ToDTO(s.baseURL)
	return &dto, nil
}

func (s *service) List(ctx context.Context, filter MediaFilter) ([]*MediaDTO, int64, error) {
	filter.Normalize()

	media, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*MediaDTO, len(media))
	for i, m := range media {
		dto := m.ToDTO(s.baseURL)
		dtos[i] = &dto
	}

	return dtos, total, nil
}

func (s *service) Update(ctx context.Context, uuid string, userID uint, isAdmin bool, req UpdateMediaRequest) (*MediaDTO, error) {
	media, err := s.repo.FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	// 权限检查
	if !isAdmin && media.UserID != userID {
		return nil, ErrForbidden
	}

	if req.Alt != nil {
		media.Alt = *req.Alt
	}
	if req.Caption != nil {
		media.Caption = *req.Caption
	}

	if err := s.repo.Update(ctx, media); err != nil {
		return nil, err
	}

	dto := media.ToDTO(s.baseURL)
	return &dto, nil
}

func (s *service) GenerateThumbnails(ctx context.Context, uuid string, userID uint, isAdmin bool, sizes []string) (*MediaDTO, error) {
	media, err := s.repo.FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	// 权限检查
	if !isAdmin && media.UserID != userID {
		return nil, ErrForbidden
	}

	// 检查是否为图片
	if !IsImage(media.MIMEType) {
		return nil, ErrNotImage
	}

	// 确定要生成的尺寸
	thumbnailSizes := make([]ThumbnailSize, 0)
	for _, name := range sizes {
		for _, def := range DefaultThumbnailSizes {
			if def.Name == name {
				thumbnailSizes = append(thumbnailSizes, def)
				break
			}
		}
	}
	if len(thumbnailSizes) == 0 {
		// 默认使用medium
		thumbnailSizes = append(thumbnailSizes, DefaultThumbnailSizes[1])
	}

	// 生成缩略图
	thumbs, err := s.processor.GenerateAndSaveThumbnails(ctx, s.storage, media.StorageKey, thumbnailSizes)
	if err != nil {
		return nil, fmt.Errorf("generate thumbnails: %w", err)
	}

	// 更新数据库
	media.SetThumbnails(thumbs)
	if err := s.repo.Update(ctx, media); err != nil {
		return nil, err
	}

	dto := media.ToDTO(s.baseURL)
	return &dto, nil
}

func (s *service) isAllowedType(contentType string) bool {
	for _, t := range s.config.AllowedTypes {
		if t == contentType {
			return true
		}
	}
	return false
}
""",

    "internal/media/handler.go": r"""package media

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/gopress/pkg/response"
)

// Handler 媒体处理程序
type Handler struct {
	svc Service
}

// NewHandler 创建媒体处理器
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterPublicRoutes 注册公开路由
func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	// 公开路由暂无
}

// RegisterAuthRoutes 注册认证用户路由
func (h *Handler) RegisterAuthRoutes(rg *gin.RouterGroup) {
	rg.POST("/media/upload", h.Upload)
	rg.GET("/media", h.List)
	rg.GET("/media/:uuid", h.Get)
	rg.DELETE("/media/:uuid", h.Delete)
	rg.PUT("/media/:uuid", h.Update)
	rg.POST("/media/:uuid/thumbnails", h.GenerateThumbnails)
}

// RegisterAdminRoutes 注册管理员路由
func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	// 管理员路由通过中间件 RequireRole("admin") 在 RegisterAuthRoutes 中统一处理
}

func getUserID(c *gin.Context) uint {
	if id, exists := c.Get("user_id"); exists {
		return id.(uint)
	}
	return 0
}

func isAdmin(c *gin.Context) bool {
	if role, exists := c.Get("user_role"); exists {
		return role == "admin"
	}
	return false
}

// Upload 上传文件
// POST /api/media/upload
func (h *Handler) Upload(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		response.Fail(c, http.StatusUnauthorized, "please login first")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "please select a file to upload")
		return
	}

	media, err := h.svc.Upload(c.Request.Context(), file, userID)
	if err != nil {
		switch err {
		case ErrInvalidFileType:
			response.Fail(c, http.StatusBadRequest, "invalid file type")
		case ErrFileTooLarge:
			response.Fail(c, http.StatusBadRequest, "file too large")
		default:
			response.Fail(c, http.StatusInternalServerError, fmt.Sprintf("upload failed: %v", err))
		}
		return
	}

	dto := media.ToDTO("")
	response.Created(c, dto)
}

// List 获取媒体列表
// GET /api/media?page=1&page_size=20&type=image&keyword=xxx
func (h *Handler) List(c *gin.Context) {
	var filter MediaFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid parameters")
		return
	}

	// 非admin只能查看自己的文件
	if !isAdmin(c) {
		userID := getUserID(c)
		filter.UserID = &userID
	}

	dtos, total, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Page(c, dtos, total, filter.Page, filter.PageSize)
}

// Get 获取单个媒体
// GET /api/media/:uuid
func (h *Handler) Get(c *gin.Context) {
	uuid := c.Param("uuid")
	userID := getUserID(c)

	dto, err := h.svc.GetByUUID(c.Request.Context(), uuid, userID, isAdmin(c))
	if err != nil {
		if err == ErrMediaNotFound {
			response.Fail(c, http.StatusNotFound, "media not found")
		} else if err == ErrForbidden {
			response.Fail(c, http.StatusForbidden, "forbidden")
		} else {
			response.Fail(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.OK(c, dto)
}

// Delete 删除媒体
// DELETE /api/media/:uuid
func (h *Handler) Delete(c *gin.Context) {
	uuid := c.Param("uuid")
	userID := getUserID(c)

	if err := h.svc.Delete(c.Request.Context(), uuid, userID, isAdmin(c)); err != nil {
		if err == ErrMediaNotFound {
			response.Fail(c, http.StatusNotFound, "media not found")
		} else if err == ErrForbidden {
			response.Fail(c, http.StatusForbidden, "forbidden")
		} else {
			response.Fail(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.OK(c, gin.H{"message": "deleted"})
}

// Update 更新媒体元信息
// PUT /api/media/:uuid
func (h *Handler) Update(c *gin.Context) {
	uuid := c.Param("uuid")
	userID := getUserID(c)

	var req UpdateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid parameters")
		return
	}

	dto, err := h.svc.Update(c.Request.Context(), uuid, userID, isAdmin(c), req)
	if err != nil {
		if err == ErrMediaNotFound {
			response.Fail(c, http.StatusNotFound, "media not found")
		} else if err == ErrForbidden {
			response.Fail(c, http.StatusForbidden, "forbidden")
		} else {
			response.Fail(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.OK(c, dto)
}

// GenerateThumbnails 生成缩略图
// POST /api/media/:uuid/thumbnails
func (h *Handler) GenerateThumbnails(c *gin.Context) {
	uuid := c.Param("uuid")
	userID := getUserID(c)

	var req GenerateThumbnailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有请求体，使用默认尺寸
		req.Sizes = []string{"medium"}
	}

	dto, err := h.svc.GenerateThumbnails(c.Request.Context(), uuid, userID, isAdmin(c), req.Sizes)
	if err != nil {
		if err == ErrMediaNotFound {
			response.Fail(c, http.StatusNotFound, "media not found")
		} else if err == ErrForbidden {
			response.Fail(c, http.StatusForbidden, "forbidden")
		} else if err == ErrNotImage {
			response.Fail(c, http.StatusBadRequest, "not an image file")
		} else {
			response.Fail(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.OK(c, dto)
}
""",

    # ========== theme 模块 ==========
    "internal/theme/model.go": r"""package theme

import (
	"time"
)

// Theme 主题数据库模型
type Theme struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Slug        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"slug"`
	Version     string    `gorm:"type:varchar(20)" json:"version"`
	Author      string    `gorm:"type:varchar(100)" json:"author"`
	Description string    `gorm:"type:text" json:"description"`
	Screenshot  string    `gorm:"type:varchar(255)" json:"screenshot"`
	IsActive    bool      `gorm:"not null;default:false" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Theme) TableName() string {
	return "themes"
}

// ThemeConfig 主题YAML配置
type ThemeConfig struct {
	Name        string `yaml:"name"`
	Slug        string `yaml:"slug"`
	Version     string `yaml:"version"`
	Author      string `yaml:"author"`
	Description string `yaml:"description"`
	Screenshot  string `yaml:"screenshot"`
}

// BaseContext 通用模板上下文
type BaseContext struct {
	SiteName    string `json:"site_name"`
	SiteURL     string `json:"site_url"`
	CurrentYear int    `json:"current_year"`
	ThemePath   string `json:"theme_path"`
	AssetsPath  string `json:"assets_path"`
}

// HomeContext 首页模板上下文
type HomeContext struct {
	BaseContext
	Posts      []PostSummary `json:"posts"`
	Pagination Pagination    `json:"pagination"`
}

// PostContext 文章详情模板上下文
type PostContext struct {
	BaseContext
	Post       PostDetail `json:"post"`
	Categories []Taxonomy `json:"categories"`
	Tags       []Taxonomy `json:"tags"`
}

// PageContext 页面模板上下文
type PageContext struct {
	BaseContext
	Page PageDetail `json:"page"`
}

// PostSummary 文章摘要（用于列表）
type PostSummary struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Excerpt     string    `json:"excerpt"`
	AuthorName  string    `json:"author_name"`
	Content     string    `json:"content"`
	ViewCount   int       `json:"view_count"`
	PublishedAt time.Time `json:"published_at"`
}

// PostDetail 文章详情
type PostDetail struct {
	ID             uint      `json:"id"`
	Title          string    `json:"title"`
	Slug           string    `json:"slug"`
	Content        string    `json:"content"`
	RawContent     string    `json:"raw_content"`
	Excerpt        string    `json:"excerpt"`
	AuthorID       uint      `json:"author_id"`
	AuthorName     string    `json:"author_name"`
	ViewCount      int       `json:"view_count"`
	CommentAllowed bool      `json:"comment_allowed"`
	PublishedAt    time.Time `json:"published_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// PageDetail 页面详情
type PageDetail struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Content     string    `json:"content"`
	RawContent  string    `json:"raw_content"`
	AuthorID    uint      `json:"author_id"`
	AuthorName  string    `json:"author_name"`
	ViewCount   int       `json:"view_count"`
	PublishedAt time.Time `json:"published_at"`
}

// Pagination 分页信息
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
	HasPrev    bool `json:"has_prev"`
	HasNext    bool `json:"has_next"`
}

// Taxonomy 分类/标签
type Taxonomy struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
""",
}

for rel, content in files.items():
    path = os.path.join(base, rel)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w", encoding="utf-8") as f:
        f.write(content)
    print("Written:", rel)

print("Done.")
