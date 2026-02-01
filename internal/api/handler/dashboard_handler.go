package handler

import (
	"net/http"

	"search-engine-go/internal/domain"
	"search-engine-go/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DashboardHandler struct {
	service *service.ContentService
	log     *zap.Logger
}

func NewDashboardHandler(service *service.ContentService, log *zap.Logger) *DashboardHandler {
	return &DashboardHandler{
		service: service,
		log:     log,
	}
}

func (h *DashboardHandler) Index(c *gin.Context) {
	var req domain.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = domain.SearchRequest{
			Page:     1,
			PageSize: 20,
			SortBy:   "score",
		}
	}

	if req.ContentType != nil {
		ct := string(*req.ContentType)
		if ct == "" || (ct != string(domain.ContentTypeVideo) && ct != string(domain.ContentTypeText)) {
			req.ContentType = nil
		}
	}

	paginationSpec := domain.NewPaginationSpecification()
	paginationSpec.NormalizePagination(&req)
	if req.SortBy == "" {
		req.SortBy = "score"
	}

	if req.SortOrder != "asc" && req.SortOrder != "desc" {
		req.SortOrder = "desc"
	}

	resp, err := h.service.Search(c.Request.Context(), &req)
	if err != nil {
		h.log.Error("Dashboard search failed", zap.Error(err))
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load content",
		})
		return
	}

	contentType := ""
	if req.ContentType != nil {
		contentType = string(*req.ContentType)
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":       "Content Search Dashboard",
		"items":       resp.Items,
		"total":       resp.Total,
		"page":        resp.Page,
		"pageSize":    resp.PageSize,
		"totalPages":  resp.TotalPages,
		"query":       req.Query,
		"sortBy":      req.SortBy,
		"sortOrder":   req.SortOrder,
		"contentType": contentType,
	})
}
