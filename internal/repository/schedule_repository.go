// Package repository encapsula o acesso ao banco de dados via GORM.
// Nenhuma outra camada deve acessar o GORM diretamente.
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/internal/domain"
	"gorm.io/gorm"
)

// ScheduleRepository define a interface para operações de persistência de agendamentos.
// Utilizar interface permite desacoplar a camada de service do GORM e facilita testes com mocks.
type ScheduleRepository interface {
	Create(ctx context.Context, schedule *domain.Schedule) error
	FindAll(ctx context.Context, filter domain.ScheduleFilter) ([]domain.Schedule, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Schedule, error)
	Update(ctx context.Context, schedule *domain.Schedule) error
	DeleteContactsByScheduleID(ctx context.Context, scheduleID uuid.UUID) error
	CreateContacts(ctx context.Context, contacts []domain.Contact) error
}

// scheduleRepository é a implementação concreta do ScheduleRepository usando GORM.
type scheduleRepository struct {
	db *gorm.DB
}

// NewScheduleRepository cria uma nova instância do repositório.
// Recebe a conexão do GORM por injeção de dependência.
func NewScheduleRepository(db *gorm.DB) ScheduleRepository {
	return &scheduleRepository{db: db}
}

// Create insere um novo agendamento com seus contatos no banco de dados.
// Utiliza transação para garantir atomicidade na criação do schedule + contacts.
func (r *scheduleRepository) Create(ctx context.Context, schedule *domain.Schedule) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(schedule).Error; err != nil {
			return err
		}
		return nil
	})
}

// FindAll busca agendamentos com suporte a filtros e paginação.
// Retorna a lista de agendamentos, o total de registros e um possível erro.
func (r *scheduleRepository) FindAll(ctx context.Context, filter domain.ScheduleFilter) ([]domain.Schedule, int64, error) {
	var schedules []domain.Schedule
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Schedule{})

	// Aplica filtro de status se fornecido
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	// Conta total antes da paginação para informações de paginação
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Aplica paginação
	offset := (filter.Page - 1) * filter.Limit
	if err := query.Preload("Contacts").
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&schedules).Error; err != nil {
		return nil, 0, err
	}

	return schedules, total, nil
}

// FindByID busca um agendamento pelo ID, incluindo seus contatos.
func (r *scheduleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
	var schedule domain.Schedule
	if err := r.db.WithContext(ctx).Preload("Contacts").First(&schedule, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &schedule, nil
}

// Update atualiza os campos de um agendamento existente.
func (r *scheduleRepository) Update(ctx context.Context, schedule *domain.Schedule) error {
	return r.db.WithContext(ctx).Save(schedule).Error
}

// DeleteContactsByScheduleID remove todos os contatos de um agendamento.
// Utilizado na atualização para substituir a lista de contatos.
func (r *scheduleRepository) DeleteContactsByScheduleID(ctx context.Context, scheduleID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("schedule_id = ?", scheduleID).Delete(&domain.Contact{}).Error
}

// CreateContacts insere uma lista de contatos no banco de dados.
func (r *scheduleRepository) CreateContacts(ctx context.Context, contacts []domain.Contact) error {
	return r.db.WithContext(ctx).Create(&contacts).Error
}
