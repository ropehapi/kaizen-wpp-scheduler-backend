// Package response fornece estruturas padronizadas para respostas HTTP.
// Todas as respostas da API seguem o mesmo formato JSON para consistência.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse é a estrutura padrão de resposta da API.
// Sempre retorna "data" e "error", onde um deles será null.
type APIResponse struct {
	Data  interface{} `json:"data"`
	Error *string     `json:"error"`
}

// PaginatedResponse estende a resposta padrão com informações de paginação.
type PaginatedResponse struct {
	Data       interface{}     `json:"data"`
	Error      *string         `json:"error"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
}

// PaginationInfo contém metadados de paginação.
type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"totalPages"`
}

// Success envia uma resposta de sucesso com os dados fornecidos.
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Data:  data,
		Error: nil,
	})
}

// SuccessWithPagination envia uma resposta de sucesso com dados e paginação.
func SuccessWithPagination(c *gin.Context, data interface{}, pagination PaginationInfo) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       data,
		Error:      nil,
		Pagination: &pagination,
	})
}

// Error envia uma resposta de erro com a mensagem fornecida.
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, APIResponse{
		Data:  nil,
		Error: &message,
	})
}

// ValidationError envia uma resposta de erro de validação (400).
func ValidationError(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// NotFoundError envia uma resposta de recurso não encontrado (404).
func NotFoundError(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalError envia uma resposta de erro interno (500).
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}
