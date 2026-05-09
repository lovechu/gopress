package content

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yourorg/gopress/internal/taxonomy"
	"github.com/yourorg/gopress/internal/user"
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
	return &dto, nil
}
