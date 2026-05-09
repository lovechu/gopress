package content

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/gopress/internal/user"
	"github.com/yourorg/gopress/pkg/response"
)

type Handler struct{ svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// RegisterPublicRoutes 注册公开路由
func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/posts", h.ListPosts)
	rg.GET("/posts/:id", h.GetPost)
	rg.GET("/pages", h.ListPages)
	rg.GET("/pages/:id", h.GetPage)
}

// RegisterAuthRoutes 注册认证用户路由
func (h *Handler) RegisterAuthRoutes(rg *gin.RouterGroup) {
	rg.POST("/posts", h.CreatePost)
	rg.PUT("/posts/:id", h.UpdatePost)
	rg.DELETE("/posts/:id", h.DeletePost)
	rg.GET("/posts/:id/revisions", h.ListRevisions)
}

// RegisterAdminRoutes 注册管理员路由
func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	// admin 权限已通过 middleware 完成
}

func getActor(c *gin.Context) (uint, user.Role) {
	u, _ := c.Get("user_id")
	r, _ := c.Get("user_role")
	return u.(uint), user.Role(r.(string))
}

// CreatePost godoc
// @Summary      创建文章
// @Description  创建新文章
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreatePostRequest true "文章信息"
// @Success      201  {object}  PostDTO
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/posts [post]
func (h *Handler) CreatePost(c *gin.Context) {
	authorID, _ := getActor(c)
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	dto, err := h.svc.Create(c.Request.Context(), authorID, req)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Created(c, dto)
}

// UpdatePost godoc
// @Summary      更新文章
// @Description  更新指定ID的文章
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int                true  "文章ID"
// @Param        request body UpdatePostRequest true "更新信息"
// @Success      200  {object}  PostDTO
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/posts/{id} [put]
func (h *Handler) UpdatePost(c *gin.Context) {
	postID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	actorID, actorRole := getActor(c)
	var req UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	dto, err := h.svc.Update(c.Request.Context(), uint(postID), actorID, actorRole, req)
	if err != nil {
		if err == ErrForbidden {
			response.Fail(c, 403, "forbidden")
		} else {
			response.Fail(c, 500, err.Error())
		}
		return
	}
	response.OK(c, dto)
}

// DeletePost godoc
// @Summary      删除文章
// @Description  删除指定ID的文章
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "文章ID"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/posts/{id} [delete]
func (h *Handler) DeletePost(c *gin.Context) {
	postID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	actorID, actorRole := getActor(c)
	if err := h.svc.Delete(c.Request.Context(), uint(postID), actorID, actorRole); err != nil {
		if err == ErrForbidden {
			response.Fail(c, 403, "forbidden")
		} else {
			response.Fail(c, 500, err.Error())
		}
		return
	}
	response.OK(c, gin.H{"message": "deleted"})
}

// GetPost godoc
// @Summary      获取文章详情
// @Description  获取指定ID的文章详情
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "文章ID"
// @Success      200  {object}  PostDTO
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/posts/{id} [get]
func (h *Handler) GetPost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	dto, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.Fail(c, 404, "post not found")
		return
	}
	response.OK(c, dto)
}

// GetPage godoc
// @Summary      获取页面详情
// @Description  获取指定ID的页面详情
// @Tags         Pages
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "页面ID"
// @Success      200  {object}  PostDTO
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/pages/{id} [get]
func (h *Handler) GetPage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	dto, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.Fail(c, 404, "page not found")
		return
	}
	response.OK(c, dto)
}

// ListPosts godoc
// @Summary      获取文章列表
// @Description  获取文章列表（支持分页和筛选）
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Param        page       query     int     false  "页码"      default(1)
// @Param        page_size  query     int     false  "每页数量"  default(20)
// @Param        status     query     string  false  "状态筛选"
// @Success      200        {object}  response.PageResponse
// @Failure      500        {object}  response.Response
// @Router       /api/posts [get]
func (h *Handler) ListPosts(c *gin.Context) { h.listContent(c, "post") }

// ListPages godoc
// @Summary      获取页面列表
// @Description  获取页面列表（支持分页和筛选）
// @Tags         Pages
// @Accept       json
// @Produce      json
// @Param        page       query     int     false  "页码"      default(1)
// @Param        page_size  query     int     false  "每页数量"  default(20)
// @Param        status     query     string  false  "状态筛选"
// @Success      200        {object}  response.PageResponse
// @Failure      500        {object}  response.Response
// @Router       /api/pages [get]
func (h *Handler) ListPages(c *gin.Context) { h.listContent(c, "page") }

func (h *Handler) listContent(c *gin.Context, contentType string) {
	filter := ListFilter{Type: contentType, Page: 1, PageSize: 20}
	if page, err := strconv.Atoi(c.Query("page")); err == nil && page > 0 {
		filter.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("page_size")); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}
	if status := c.Query("status"); status != "" {
		filter.Status = status
	}
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
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Page(c, dtos, total, filter.Page, filter.PageSize)
}

// ListRevisions godoc
// @Summary      获取文章修订版本列表
// @Description  获取指定文章的修订版本列表
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "文章ID"
// @Success      200  {array}   RevisionDTO
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/posts/{id}/revisions [get]
func (h *Handler) ListRevisions(c *gin.Context) {
	postID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	revs, err := h.svc.ListRevisions(c.Request.Context(), uint(postID))
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, revs)
}
