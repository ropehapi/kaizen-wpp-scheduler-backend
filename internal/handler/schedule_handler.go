// Package handler contém os controllers HTTP que recebem as requisições,
// fazem validação de input e delegam a lógica para a camada de service.
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/internal/domain"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/internal/service"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/pkg/response"
)

// ScheduleHandler encapsula os handlers de agendamento.
type ScheduleHandler struct {
	service service.ScheduleService
}

// NewScheduleHandler cria uma nova instância do handler.
func NewScheduleHandler(svc service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{service: svc}
}

// CreateSchedule godoc
// @Summary Cria um novo agendamento
// @Description Cria um agendamento com mensagem, contatos, tipo e data
// @Accept json
// @Produce json
// @Param schedule body domain.CreateScheduleRequest true "Dados do agendamento"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/schedules [post]
func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	var req domain.CreateScheduleRequest

	// Validação de input usando binding do Gin
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Dados inválidos: "+err.Error())
		return
	}

	schedule, err := h.service.CreateSchedule(c.Request.Context(), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, schedule)
}

// ListSchedules godoc
// @Summary Lista agendamentos
// @Description Retorna agendamentos com filtros e paginação
// @Produce json
// @Param status query string false "Filtro por status"
// @Param page query int false "Página"
// @Param limit query int false "Limite por página"
// @Success 200 {object} response.PaginatedResponse
// @Router /api/v1/schedules [get]
func (h *ScheduleHandler) ListSchedules(c *gin.Context) {
	filter := domain.ScheduleFilter{}

	// Parse do filtro de status (query param opcional)
	if status := c.Query("status"); status != "" {
		s := domain.ScheduleStatus(status)
		// Valida se o status é válido
		if s != domain.ScheduleStatusScheduled && s != domain.ScheduleStatusSent && s != domain.ScheduleStatusCanceled {
			response.ValidationError(c, "Status inválido. Valores aceitos: scheduled, sent, canceled")
			return
		}
		filter.Status = &s
	}

	// Parse de paginação
	if page := c.Query("page"); page != "" {
		p, err := strconv.Atoi(page)
		if err != nil || p < 1 {
			response.ValidationError(c, "Parâmetro 'page' inválido")
			return
		}
		filter.Page = p
	}

	if limit := c.Query("limit"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil || l < 1 {
			response.ValidationError(c, "Parâmetro 'limit' inválido")
			return
		}
		filter.Limit = l
	}

	schedules, total, err := h.service.ListSchedules(c.Request.Context(), filter)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Calcula total de páginas
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	totalPages := total / int64(limit)
	if total%int64(limit) != 0 {
		totalPages++
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}

	response.SuccessWithPagination(c, schedules, response.PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetScheduleByID godoc
// @Summary Busca agendamento por ID
// @Description Retorna um agendamento específico pelo UUID
// @Produce json
// @Param id path string true "ID do agendamento"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/schedules/{id} [get]
func (h *ScheduleHandler) GetScheduleByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "ID inválido: deve ser um UUID válido")
		return
	}

	schedule, err := h.service.GetScheduleByID(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, schedule)
}

// UpdateSchedule godoc
// @Summary Atualiza um agendamento
// @Description Atualiza campos de um agendamento existente
// @Accept json
// @Produce json
// @Param id path string true "ID do agendamento"
// @Param schedule body domain.UpdateScheduleRequest true "Dados para atualização"
// @Success 200 {object} response.APIResponse
// @Failure 400,404 {object} response.APIResponse
// @Router /api/v1/schedules/{id} [put]
func (h *ScheduleHandler) UpdateSchedule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "ID inválido: deve ser um UUID válido")
		return
	}

	var req domain.UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Dados inválidos: "+err.Error())
		return
	}

	schedule, err := h.service.UpdateSchedule(c.Request.Context(), id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, schedule)
}

// CancelSchedule godoc
// @Summary Cancela um agendamento
// @Description Altera o status de um agendamento para "canceled"
// @Produce json
// @Param id path string true "ID do agendamento"
// @Success 200 {object} response.APIResponse
// @Failure 400,404 {object} response.APIResponse
// @Router /api/v1/schedules/{id}/cancel [patch]
func (h *ScheduleHandler) CancelSchedule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "ID inválido: deve ser um UUID válido")
		return
	}

	schedule, err := h.service.CancelSchedule(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, schedule)
}

// handleServiceError traduz erros de serviço em respostas HTTP adequadas.
// Centraliza o tratamento de erros para evitar duplicação.
func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrScheduleNotFound):
		response.NotFoundError(c, err.Error())
	case errors.Is(err, service.ErrScheduleAlreadySent):
		response.Error(c, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrFrequencyRequired):
		response.ValidationError(c, err.Error())
	case errors.Is(err, service.ErrInvalidFrequency):
		response.ValidationError(c, err.Error())
	case errors.Is(err, service.ErrAlreadyCanceled):
		response.Error(c, http.StatusConflict, err.Error())
	default:
		response.InternalError(c, "Erro interno do servidor")
	}
}
