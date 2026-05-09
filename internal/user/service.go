package user

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
