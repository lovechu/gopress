package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/gopress/pkg/response"
)

type Handler struct{ svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// RegisterPublicRoutes 注册公开路由
func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.POST("/auth/register", h.Register)
	rg.POST("/auth/login", h.Login)
	rg.POST("/auth/refresh", h.RefreshToken)
}

// RegisterAuthRoutes 注册认证用户路由
func (h *Handler) RegisterAuthRoutes(rg *gin.RouterGroup) {
	rg.GET("/users/me", h.GetProfile)
	rg.PUT("/users/me", h.UpdateProfile)
	rg.POST("/users/me/password", h.ChangePassword)
}

// RegisterAdminRoutes 注册管理员路由
func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/users", h.ListUsers)
}

// Register godoc
// @Summary      用户注册
// @Description  注册新用户账号
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "注册信息"
// @Success      201  {object}  UserDTO
// @Failure      400  {object}  response.Response
// @Failure      409  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	dto, err := h.svc.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrDuplicateUsername) {
			response.Fail(c, http.StatusConflict, "username already exists")
		} else if errors.Is(err, ErrDuplicateEmail) {
			response.Fail(c, http.StatusConflict, "email already exists")
		} else {
			response.Fail(c, http.StatusInternalServerError, "registration failed")
		}
		return
	}
	response.Created(c, dto)
}

// Login godoc
// @Summary      用户登录
// @Description  使用用户名和密码登录
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "登录信息"
// @Success      200  {object}  TokenPair
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	pair, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Fail(c, http.StatusUnauthorized, "invalid credentials")
		} else if errors.Is(err, ErrUserInactive) {
			response.Fail(c, http.StatusForbidden, "account is disabled")
		} else {
			response.Fail(c, http.StatusInternalServerError, "login failed")
		}
		return
	}
	response.OK(c, pair)
}

// RefreshToken godoc
// @Summary      刷新Token
// @Description  使用refresh token获取新的access token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body RefreshRequest true "Refresh Token"
// @Success      200  {object}  TokenPair
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	pair, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}
	response.OK(c, pair)
}

// GetProfile godoc
// @Summary      获取个人资料
// @Description  获取当前登录用户的个人资料
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  UserDTO
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/users/me [get]
func (h *Handler) GetProfile(c *gin.Context) {
	u, _ := c.Get("user_id")
	userID := u.(uint)
	dto, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.Fail(c, http.StatusNotFound, "user not found")
		} else {
			response.Fail(c, http.StatusInternalServerError, "failed to get profile")
		}
		return
	}
	response.OK(c, dto)
}

// UpdateProfile godoc
// @Summary      更新个人资料
// @Description  更新当前登录用户的个人资料
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body UpdateProfileRequest true "更新信息"
// @Success      200  {object}  UserDTO
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/users/me [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	u, _ := c.Get("user_id")
	userID := u.(uint)
	dto, err := h.svc.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "update failed")
		return
	}
	response.OK(c, dto)
}

// ChangePassword godoc
// @Summary      修改密码
// @Description  修改当前登录用户的密码
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body ChangePasswordRequest true "密码信息"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/users/me/password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	u, _ := c.Get("user_id")
	userID := u.(uint)
	if err := h.svc.ChangePassword(c.Request.Context(), userID, req); err != nil {
		if errors.Is(err, ErrWrongPassword) {
			response.Fail(c, http.StatusUnauthorized, "wrong password")
		} else {
			response.Fail(c, http.StatusInternalServerError, "change password failed")
		}
		return
	}
	response.OK(c, gin.H{"message": "password changed"})
}

// ListUsers godoc
// @Summary      获取用户列表
// @Description  获取所有用户列表（管理员）
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "页码" default(1)
// @Param        page_size query int false "每页数量" default(20)
// @Success      200  {object}  response.PageResponse
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	dtos, total, err := h.svc.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "failed to list users")
		return
	}
	response.Page(c, dtos, total, page, pageSize)
}

func mustGetUserID(c *gin.Context) uint {
	u, _ := c.Get("user_id")
	return u.(uint)
}
