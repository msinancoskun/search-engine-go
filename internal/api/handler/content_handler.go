package handler

import (
	"net/http"
	"strconv"

	"search-engine-go/internal/api/middleware"
	"search-engine-go/internal/domain"
	"search-engine-go/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ContentHandler struct {
	service service.ContentServiceInterface
	log     *zap.Logger
}

func NewContentHandler(service service.ContentServiceInterface, log *zap.Logger) *ContentHandler {
	return &ContentHandler{
		service: service,
		log:     log,
	}
}

func (h *ContentHandler) Search(c *gin.Context) {
	var req domain.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Warn("Invalid search request", zap.Error(err), zap.String("request_id", middleware.GetRequestID(c)))
		domainErr := domain.NewInvalidInputError("query", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     domainErr.Message,
			"code":      string(domainErr.Code),
			"details":   domainErr.Details,
			"request_id": middleware.GetRequestID(c),
		})
		return
	}

	if req.ContentType != nil {
		ct := string(*req.ContentType)
		if ct == "" || (ct != string(domain.ContentTypeVideo) && ct != string(domain.ContentTypeText)) {
			req.ContentType = nil
		}
	}

	paginationSpec := domain.NewPaginationSpecification()
	paginationSpec.NormalizePagination(&req)

	resp, err := h.service.Search(c.Request.Context(), &req)
	if err != nil {
		requestID := middleware.GetRequestID(c)
		h.log.Error("Search failed", zap.Error(err), zap.String("request_id", requestID))
		
		if domainErr, ok := err.(*domain.DomainError); ok {
			statusCode := http.StatusInternalServerError
			if domainErr.Code == domain.ErrorCodeInvalidInput {
				statusCode = http.StatusBadRequest
			} else if domainErr.Code == domain.ErrorCodeProviderError {
				// Provider errors might be partial failures, return 503 or 500
				statusCode = http.StatusServiceUnavailable
			}
			c.JSON(statusCode, gin.H{
				"error":      domainErr.Message,
				"code":       string(domainErr.Code),
				"details":    domainErr.Details,
				"request_id": requestID,
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal server error",
			"request_id": requestID,
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ContentHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		requestID := middleware.GetRequestID(c)
		domainErr := domain.NewInvalidInputError("id", "must be a valid integer")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      domainErr.Message,
			"code":       string(domainErr.Code),
			"details":    domainErr.Details,
			"request_id": requestID,
		})
		return
	}

	content, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		requestID := middleware.GetRequestID(c)
		h.log.Error("Get content failed", zap.Error(err), zap.String("request_id", requestID))
		
		if domainErr, ok := err.(*domain.DomainError); ok {
			statusCode := http.StatusNotFound
			if domainErr.Code == domain.ErrorCodeInvalidInput {
				statusCode = http.StatusBadRequest
			}
			c.JSON(statusCode, gin.H{
				"error":      domainErr.Message,
				"code":       string(domainErr.Code),
				"details":    domainErr.Details,
				"request_id": requestID,
			})
			return
		}
		
		if err.Error() != "" && (err.Error() == "record not found" || 
			err.Error() == "sql: no rows in result set") {
			domainErr := domain.NewNotFoundError("content", id)
			c.JSON(http.StatusNotFound, gin.H{
				"error":      domainErr.Message,
				"code":       string(domainErr.Code),
				"details":    domainErr.Details,
				"request_id": requestID,
			})
			return
		}
		
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "Content not found",
			"request_id": requestID,
		})
		return
	}

	c.JSON(http.StatusOK, content)
}
