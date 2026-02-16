// Package domain contém os modelos de domínio da aplicação.
// Os modelos são agnósticos de framework e representam as entidades do negócio.
package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ScheduleType define os tipos possíveis de agendamento.
type ScheduleType string

const (
	ScheduleTypeOnce      ScheduleType = "once"
	ScheduleTypeRecurring ScheduleType = "recurring"
)

// ScheduleStatus define os status possíveis de um agendamento.
type ScheduleStatus string

const (
	ScheduleStatusScheduled ScheduleStatus = "scheduled"
	ScheduleStatusSent      ScheduleStatus = "sent"
	ScheduleStatusCanceled  ScheduleStatus = "canceled"
)

// Frequency define as frequências possíveis para agendamentos recorrentes.
type Frequency string

const (
	FrequencyDaily   Frequency = "daily"
	FrequencyWeekly  Frequency = "weekly"
	FrequencyMonthly Frequency = "monthly"
)

// Schedule representa um agendamento de envio de mensagens em lote.
// Cada agendamento contém uma mensagem, tipo de envio, e uma lista de contatos.
type Schedule struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key"`
	Message     string         `json:"message" gorm:"type:text;not null"`
	Type        ScheduleType   `json:"type" gorm:"type:varchar(20);not null"`
	Frequency   *Frequency     `json:"frequency" gorm:"type:varchar(20)"`
	ScheduledAt time.Time      `json:"scheduledAt" gorm:"not null"`
	Status      ScheduleStatus `json:"status" gorm:"type:varchar(20);not null;default:'scheduled'"`
	Contacts    []Contact      `json:"contacts" gorm:"foreignKey:ScheduleID;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

// BeforeCreate é um hook do GORM que gera um UUID antes de inserir o registro.
func (s *Schedule) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// Contact representa um contato associado a um agendamento.
type Contact struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	ScheduleID uuid.UUID `json:"scheduleId" gorm:"type:uuid;not null;index"`
	Name       string    `json:"name" gorm:"type:varchar(255);not null"`
	Phone      string    `json:"phone" gorm:"type:varchar(20);not null"`
}

// BeforeCreate é um hook do GORM que gera um UUID antes de inserir o registro.
func (c *Contact) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// CreateScheduleRequest representa o payload de criação de agendamento.
// Utilizado na camada de handler para fazer bind do JSON recebido.
type CreateScheduleRequest struct {
	Message     string                `json:"message" binding:"required"`
	Type        ScheduleType          `json:"type" binding:"required,oneof=once recurring"`
	Frequency   *Frequency            `json:"frequency"`
	ScheduledAt time.Time             `json:"scheduledAt" binding:"required"`
	Contacts    []CreateContactRequest `json:"contacts" binding:"required,min=1,dive"`
}

// CreateContactRequest representa o payload de um contato na criação.
type CreateContactRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

// UpdateScheduleRequest representa o payload de atualização de agendamento.
type UpdateScheduleRequest struct {
	Message     *string      `json:"message"`
	Type        *ScheduleType `json:"type" binding:"omitempty,oneof=once recurring"`
	Frequency   *Frequency   `json:"frequency"`
	ScheduledAt *time.Time   `json:"scheduledAt"`
	Contacts    []CreateContactRequest `json:"contacts" binding:"omitempty,min=1,dive"`
}

// ScheduleFilter contém os filtros para listagem de agendamentos.
type ScheduleFilter struct {
	Status *ScheduleStatus
	Page   int
	Limit  int
}
