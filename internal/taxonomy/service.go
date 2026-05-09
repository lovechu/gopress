package taxonomy

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

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) Create(ctx context.Context, req CreateTermRequest) (*TermDTO, error) {
	if req.Taxonomy != "category" && req.Taxonomy != "tag" {
		return nil, ErrInvalidTaxonomy
	}
	term := &Term{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Taxonomy:    req.Taxonomy,
		ParentID:    req.ParentID,
	}
	if err := s.repo.Create(ctx, term); err != nil {
		return nil, fmt.Errorf("create term: %w", err)
	}
	dto := term.ToDTO()
	return &dto, nil
}

func (s *service) Update(ctx context.Context, id uint, req UpdateTermRequest) (*TermDTO, error) {
	term, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		term.Name = *req.Name
	}
	if req.Slug != nil {
		term.Slug = *req.Slug
	}
	if req.Description != nil {
		term.Description = *req.Description
	}
	if req.ParentID != nil {
		term.ParentID = req.ParentID
	}
	if err := s.repo.Update(ctx, term); err != nil {
		return nil, err
	}
	dto := term.ToDTO()
	return &dto, nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) GetByID(ctx context.Context, id uint) (*TermDTO, error) {
	term, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := term.ToDTO()
	return &dto, nil
}

func (s *service) List(ctx context.Context, taxonomy string, page, pageSize int) ([]*TermDTO, int64, error) {
	terms, total, err := s.repo.List(ctx, taxonomy, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	dtos := make([]*TermDTO, len(terms))
	for i, t := range terms {
		dto := t.ToDTO()
		dtos[i] = &dto
	}
	return dtos, total, nil
}
