// Package service contém a lógica de negócio da aplicação.
// Esta camada orquestra as operações e aplica validações de regras de negócio,
// delegando a persistência para a camada de repository.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/internal/domain"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/internal/repository"
	"gorm.io/gorm"
)

// Erros de negócio customizados para facilitar o tratamento na camada de handler.
var (
	ErrScheduleNotFound     = errors.New("agendamento não encontrado")
	ErrScheduleAlreadySent  = errors.New("não é possível alterar um agendamento já enviado")
	ErrFrequencyRequired    = errors.New("frequência é obrigatória para agendamentos recorrentes")
	ErrInvalidFrequency     = errors.New("frequência inválida. Valores aceitos: daily, weekly, monthly")
	ErrAlreadyCanceled      = errors.New("agendamento já está cancelado")
)

// ScheduleService define a interface da camada de serviço.
type ScheduleService interface {
	CreateSchedule(ctx context.Context, req domain.CreateScheduleRequest) (*domain.Schedule, error)
	ListSchedules(ctx context.Context, filter domain.ScheduleFilter) ([]domain.Schedule, int64, error)
	GetScheduleByID(ctx context.Context, id uuid.UUID) (*domain.Schedule, error)
	UpdateSchedule(ctx context.Context, id uuid.UUID, req domain.UpdateScheduleRequest) (*domain.Schedule, error)
	CancelSchedule(ctx context.Context, id uuid.UUID) (*domain.Schedule, error)
}

// scheduleService é a implementação concreta do ScheduleService.
type scheduleService struct {
	repo repository.ScheduleRepository
}

// NewScheduleService cria uma nova instância do serviço.
// Recebe o repositório por injeção de dependência.
func NewScheduleService(repo repository.ScheduleRepository) ScheduleService {
	return &scheduleService{repo: repo}
}

// CreateSchedule cria um novo agendamento aplicando as regras de negócio:
// - Se type = "recurring", frequency é obrigatório
// - Status inicial sempre "scheduled"
// - Contatos são associados ao agendamento via transação
func (s *scheduleService) CreateSchedule(ctx context.Context, req domain.CreateScheduleRequest) (*domain.Schedule, error) {
	// Valida regra: agendamento recorrente exige frequência
	if req.Type == domain.ScheduleTypeRecurring {
		if req.Frequency == nil {
			return nil, ErrFrequencyRequired
		}
		if !isValidFrequency(*req.Frequency) {
			return nil, ErrInvalidFrequency
		}
	}

	// Monta a entidade de domínio
	schedule := &domain.Schedule{
		Message:     req.Message,
		Type:        req.Type,
		Frequency:   req.Frequency,
		ScheduledAt: req.ScheduledAt,
		Status:      domain.ScheduleStatusScheduled,
	}

	// Converte os contatos do request para entidades de domínio
	for _, c := range req.Contacts {
		schedule.Contacts = append(schedule.Contacts, domain.Contact{
			Name:  c.Name,
			Phone: c.Phone,
		})
	}

	// Persiste no banco (repository cuida da transação)
	if err := s.repo.Create(ctx, schedule); err != nil {
		return nil, fmt.Errorf("erro ao criar agendamento: %w", err)
	}

	return schedule, nil
}

// ListSchedules retorna agendamentos com suporte a filtros e paginação.
// Valores padrão: page=1, limit=10.
func (s *scheduleService) ListSchedules(ctx context.Context, filter domain.ScheduleFilter) ([]domain.Schedule, int64, error) {
	// Aplica valores padrão de paginação
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return s.repo.FindAll(ctx, filter)
}

// GetScheduleByID busca um agendamento específico pelo ID.
func (s *scheduleService) GetScheduleByID(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
	schedule, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, fmt.Errorf("erro ao buscar agendamento: %w", err)
	}
	return schedule, nil
}

// UpdateSchedule atualiza um agendamento existente.
// Regra de negócio: não permite atualização se status = "sent".
func (s *scheduleService) UpdateSchedule(ctx context.Context, id uuid.UUID, req domain.UpdateScheduleRequest) (*domain.Schedule, error) {
	schedule, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, fmt.Errorf("erro ao buscar agendamento: %w", err)
	}

	// Regra: não permite alterar agendamento já enviado
	if schedule.Status == domain.ScheduleStatusSent {
		return nil, ErrScheduleAlreadySent
	}

	// Atualiza campos se fornecidos
	if req.Message != nil {
		schedule.Message = *req.Message
	}
	if req.Type != nil {
		schedule.Type = *req.Type
	}
	if req.Frequency != nil {
		if !isValidFrequency(*req.Frequency) {
			return nil, ErrInvalidFrequency
		}
		schedule.Frequency = req.Frequency
	}
	if req.ScheduledAt != nil {
		schedule.ScheduledAt = *req.ScheduledAt
	}

	// Valida regra de recorrência após atualização
	if schedule.Type == domain.ScheduleTypeRecurring && schedule.Frequency == nil {
		return nil, ErrFrequencyRequired
	}

	// Atualiza o agendamento
	if err := s.repo.Update(ctx, schedule); err != nil {
		return nil, fmt.Errorf("erro ao atualizar agendamento: %w", err)
	}

	// Se contatos foram fornecidos, substitui a lista existente
	if req.Contacts != nil && len(req.Contacts) > 0 {
		// Remove contatos antigos
		if err := s.repo.DeleteContactsByScheduleID(ctx, schedule.ID); err != nil {
			return nil, fmt.Errorf("erro ao remover contatos antigos: %w", err)
		}

		// Cria novos contatos
		var contacts []domain.Contact
		for _, c := range req.Contacts {
			contacts = append(contacts, domain.Contact{
				ScheduleID: schedule.ID,
				Name:       c.Name,
				Phone:      c.Phone,
			})
		}
		if err := s.repo.CreateContacts(ctx, contacts); err != nil {
			return nil, fmt.Errorf("erro ao criar novos contatos: %w", err)
		}
		schedule.Contacts = contacts
	}

	return schedule, nil
}

// CancelSchedule altera o status de um agendamento para "canceled".
func (s *scheduleService) CancelSchedule(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
	schedule, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrScheduleNotFound
		}
		return nil, fmt.Errorf("erro ao buscar agendamento: %w", err)
	}

	// Regra: não permite cancelar agendamento já enviado
	if schedule.Status == domain.ScheduleStatusSent {
		return nil, ErrScheduleAlreadySent
	}

	// Regra: não permite cancelar agendamento já cancelado
	if schedule.Status == domain.ScheduleStatusCanceled {
		return nil, ErrAlreadyCanceled
	}

	schedule.Status = domain.ScheduleStatusCanceled

	if err := s.repo.Update(ctx, schedule); err != nil {
		return nil, fmt.Errorf("erro ao cancelar agendamento: %w", err)
	}

	return schedule, nil
}

// isValidFrequency verifica se a frequência informada é válida.
func isValidFrequency(f domain.Frequency) bool {
	switch f {
	case domain.FrequencyDaily, domain.FrequencyWeekly, domain.FrequencyMonthly:
		return true
	}
	return false
}
