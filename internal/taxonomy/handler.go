package taxonomy

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/gopress/pkg/response"
)

type Handler struct{ svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// RegisterPublicRoutes 注册公开路由
func (h *Handler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/terms", h.ListTerms)
	rg.GET("/terms/:id", h.GetTerm)
}

// RegisterAdminRoutes 注册管理员路由
func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.POST("/terms", h.CreateTerm)
	rg.PUT("/terms/:id", h.UpdateTerm)
	rg.DELETE("/terms/:id", h.DeleteTerm)
}

// CreateTerm godoc
// @Summary      创建分类/标签
// @Description  创建新的分类或标签
// @Tags         Taxonomy
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateTermRequest true "分类/标签信息"
// @Success      201  {object}  TermDTO
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/terms [post]
func (h *Handler) CreateTerm(c *gin.Context) {
	var req CreateTermRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	dto, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Created(c, dto)
}

// UpdateTerm godoc
// @Summary      更新分类/标签
// @Description  更新指定ID的分类或标签
// @Tags         Taxonomy
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int                      true  "分类/标签ID"
// @Param        request body UpdateTermRequest true     "更新信息"
// @Success      200  {object}  TermDTO
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/terms/{id} [put]
func (h *Handler) UpdateTerm(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req UpdateTermRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	dto, err := h.svc.Update(c.Request.Context(), uint(id), req)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, dto)
}

// DeleteTerm godoc
// @Summary      删除分类/标签
// @Description  删除指定ID的分类或标签
// @Tags         Taxonomy
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "分类/标签ID"
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/terms/{id} [delete]
func (h *Handler) DeleteTerm(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "deleted"})
}

// GetTerm godoc
// @Summary      获取分类/标签详情
// @Description  获取指定ID的分类或标签详情
// @Tags         Taxonomy
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "分类/标签ID"
// @Success      200  {object}  TermDTO
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/terms/{id} [get]
func (h *Handler) GetTerm(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	dto, err := h.svc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.Fail(c, 404, "term not found")
		return
	}
	response.OK(c, dto)
}

// ListTerms godoc
// @Summary      获取分类/标签列表
// @Description  获取分类或标签列表（支持分页和筛选）
// @Tags         Taxonomy
// @Accept       json
// @Produce      json
// @Param        taxonomy   query     string  false  "类型筛选（category/tag）"
// @Param        page       query     int     false  "页码"      default(1)
// @Param        page_size  query     int     false  "每页数量"  default(20)
// @Success      200        {object}  response.PageResponse
// @Failure      500        {object}  response.Response
// @Router       /api/terms [get]
func (h *Handler) ListTerms(c *gin.Context) {
	taxonomy := c.Query("taxonomy")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	dtos, total, err := h.svc.List(c.Request.Context(), taxonomy, page, pageSize)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.Page(c, dtos, total, page, pageSize)
}
