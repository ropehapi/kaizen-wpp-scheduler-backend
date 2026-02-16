package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockScheduleRepository é o mock do repositório para testes unitários.
// Permite testar a camada de service sem depender do banco de dados.
type MockScheduleRepository struct {
	mock.Mock
}

func (m *MockScheduleRepository) Create(ctx context.Context, schedule *domain.Schedule) error {
	args := m.Called(ctx, schedule)
	return args.Error(0)
}

func (m *MockScheduleRepository) FindAll(ctx context.Context, filter domain.ScheduleFilter) ([]domain.Schedule, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domain.Schedule), args.Get(1).(int64), args.Error(2)
}

func (m *MockScheduleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Schedule), args.Error(1)
}

func (m *MockScheduleRepository) Update(ctx context.Context, schedule *domain.Schedule) error {
	args := m.Called(ctx, schedule)
	return args.Error(0)
}

func (m *MockScheduleRepository) DeleteContactsByScheduleID(ctx context.Context, scheduleID uuid.UUID) error {
	args := m.Called(ctx, scheduleID)
	return args.Error(0)
}

func (m *MockScheduleRepository) CreateContacts(ctx context.Context, contacts []domain.Contact) error {
	args := m.Called(ctx, contacts)
	return args.Error(0)
}

// ===== Testes de CreateSchedule =====

func TestCreateSchedule_Success_Once(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := NewScheduleService(mockRepo)

	req := domain.CreateScheduleRequest{
		Message:     "Olá, mundo!",
		Type:        domain.ScheduleTypeOnce,
		ScheduledAt: time.Now().Add(24 * time.Hour),
		Contacts: []domain.CreateContactRequest{
			{Name: "João", Phone: "+5511999999999"},
		},
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Schedule")).
		Return(nil).
		Run(func(args mock.Arguments) {
			// Simula a geração de ID pelo banco
			schedule := args.Get(1).(*domain.Schedule)
			schedule.ID = uuid.New()
		})

	schedule, err := svc.CreateSchedule(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, schedule)
	assert.Equal(t, "Olá, mundo!", schedule.Message)
	assert.Equal(t, domain.ScheduleTypeOnce, schedule.Type)
	assert.Equal(t, domain.ScheduleStatusScheduled, schedule.Status)
	assert.Len(t, schedule.Contacts, 1)
	mockRepo.AssertExpectations(t)
}

func TestCreateSchedule_Success_Recurring(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := NewScheduleService(mockRepo)

	freq := domain.FrequencyWeekly
	req := domain.CreateScheduleRequest{
		Message:     "Lembrete semanal",
		Type:        domain.ScheduleTypeRecurring,
		Frequency:   &freq,
		ScheduledAt: time.Now().Add(24 * time.Hour),
		Contacts: []domain.CreateContactRequest{
			{Name: "Maria", Phone: "+5511888888888"},
		},
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Schedule")).
		Return(nil)

	schedule, err := svc.CreateSchedule(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, schedule)
	assert.Equal(t, domain.ScheduleTypeRecurring, schedule.Type)
	assert.Equal(t, &freq, schedule.Frequency)
	mockRepo.AssertExpectations(t)
}

func TestCreateSchedule_Error_RecurringWithoutFrequency(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := NewScheduleService(mockRepo)

	req := domain.CreateScheduleRequest{
		Message:     "Sem frequência",
		Type:        domain.ScheduleTypeRecurring,
		ScheduledAt: time.Now().Add(24 * time.Hour),
		Contacts: []domain.CreateContactRequest{
			{Name: "João", Phone: "+5511999999999"},
		},
	}

	schedule, err := svc.CreateSchedule(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, schedule)
	assert.True(t, errors.Is(err, ErrFrequencyRequired))
}

func TestCreateSchedule_Error_InvalidFrequency(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := NewScheduleService(mockRepo)

	invalidFreq := domain.Frequency("hourly")
	req := domain.CreateScheduleRequest{
		Message:     "Frequência inválida",
		Type:        domain.ScheduleTypeRecurring,
		Frequency:   &invalidFreq,
		ScheduledAt: time.Now().Add(24 * time.Hour),
		Contacts: []domain.CreateContactRequest{
			{Name: "João", Phone: "+5511999999999"},
		},
	}

	schedule, err := svc.CreateSchedule(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, schedule)
	assert.True(t, errors.Is(err, ErrInvalidFrequency))
}

// ===== Testes de GetScheduleByID =====

func TestGetScheduleByID_Success(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := NewScheduleService(mockRepo)

	scheduleID := uuid.New()
	expectedSchedule := &domain.Schedule{
		ID:      scheduleID,
		Message: "Teste",
		Status:  domain.ScheduleStatusScheduled,
	}

	mockRepo.On("FindByID", mock.Anything, scheduleID).Return(expectedSchedule, nil)

	schedule, err := svc.GetScheduleByID(context.Background(), scheduleID)

	assert.NoError(t, err)
	assert.Equal(t, expectedSchedule, schedule)
	mockRepo.AssertExpectations(t)
}

func TestGetScheduleByID_NotFound(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := NewScheduleService(mockRepo)

	scheduleID := uuid.New()
	mockRepo.On("FindByID", mock.Anything, scheduleID).Return(nil, gorm.ErrRecordNotFound)

	schedule, err := svc.GetScheduleByID(context.Background(), scheduleID)

	assert.Error(t, err)
	assert.Nil(t, schedule)
	assert.True(t, errors.Is(err, ErrScheduleNotFound))
	mockRepo.AssertExpectations(t)
}

// ===== Testes de CancelSchedule =====

func TestCancelSchedule_Success(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := NewScheduleService(mockRepo)

	scheduleID := uuid.New()
	existingSchedule := &domain.Schedule{
		ID:      scheduleID,
		Message: "Teste",
		Status:  domain.ScheduleStatusScheduled,
	}

	mockRepo.On("FindByID", mock.Anything, scheduleID).Return(existingSchedule, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Schedule")).Return(nil)

	schedule, err := svc.CancelSchedule(context.Background(), scheduleID)

	assert.NoError(t, err)
	assert.Equal(t, domain.ScheduleStatusCanceled, schedule.Status)
	mockRepo.AssertExpectations(t)
}

func TestCancelSchedule_Error_AlreadySent(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := NewScheduleService(mockRepo)

	scheduleID := uuid.New()
	existingSchedule := &domain.Schedule{
		ID:     scheduleID,
		Status: domain.ScheduleStatusSent,
	}

	mockRepo.On("FindByID", mock.Anything, scheduleID).Return(existingSchedule, nil)

	schedule, err := svc.CancelSchedule(context.Background(), scheduleID)

	assert.Error(t, err)
	assert.Nil(t, schedule)
	assert.True(t, errors.Is(err, ErrScheduleAlreadySent))
}

func TestCancelSchedule_Error_AlreadyCanceled(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := NewScheduleService(mockRepo)

	scheduleID := uuid.New()
	existingSchedule := &domain.Schedule{
		ID:     scheduleID,
		Status: domain.ScheduleStatusCanceled,
	}

	mockRepo.On("FindByID", mock.Anything, scheduleID).Return(existingSchedule, nil)

	schedule, err := svc.CancelSchedule(context.Background(), scheduleID)

	assert.Error(t, err)
	assert.Nil(t, schedule)
	assert.True(t, errors.Is(err, ErrAlreadyCanceled))
}

// ===== Testes de UpdateSchedule =====

func TestUpdateSchedule_Success(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := NewScheduleService(mockRepo)

	scheduleID := uuid.New()
	existingSchedule := &domain.Schedule{
		ID:      scheduleID,
		Message: "Mensagem antiga",
		Type:    domain.ScheduleTypeOnce,
		Status:  domain.ScheduleStatusScheduled,
	}

	newMessage := "Mensagem nova"
	req := domain.UpdateScheduleRequest{
		Message: &newMessage,
	}

	mockRepo.On("FindByID", mock.Anything, scheduleID).Return(existingSchedule, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Schedule")).Return(nil)

	schedule, err := svc.UpdateSchedule(context.Background(), scheduleID, req)

	assert.NoError(t, err)
	assert.Equal(t, "Mensagem nova", schedule.Message)
	mockRepo.AssertExpectations(t)
}

func TestUpdateSchedule_Error_AlreadySent(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := NewScheduleService(mockRepo)

	scheduleID := uuid.New()
	existingSchedule := &domain.Schedule{
		ID:     scheduleID,
		Status: domain.ScheduleStatusSent,
	}

	newMessage := "Mensagem nova"
	req := domain.UpdateScheduleRequest{
		Message: &newMessage,
	}

	mockRepo.On("FindByID", mock.Anything, scheduleID).Return(existingSchedule, nil)

	schedule, err := svc.UpdateSchedule(context.Background(), scheduleID, req)

	assert.Error(t, err)
	assert.Nil(t, schedule)
	assert.True(t, errors.Is(err, ErrScheduleAlreadySent))
}

// ===== Testes de ListSchedules =====

func TestListSchedules_Success(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := NewScheduleService(mockRepo)

	expectedSchedules := []domain.Schedule{
		{ID: uuid.New(), Message: "Teste 1", Status: domain.ScheduleStatusScheduled},
		{ID: uuid.New(), Message: "Teste 2", Status: domain.ScheduleStatusScheduled},
	}

	filter := domain.ScheduleFilter{Page: 1, Limit: 10}
	mockRepo.On("FindAll", mock.Anything, filter).Return(expectedSchedules, int64(2), nil)

	schedules, total, err := svc.ListSchedules(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, schedules, 2)
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestListSchedules_DefaultPagination(t *testing.T) {
	mockRepo := new(MockScheduleRepository)
	svc := NewScheduleService(mockRepo)

	// Filtro com valores padrão (page=0, limit=0) deve ser normalizado
	expectedFilter := domain.ScheduleFilter{Page: 1, Limit: 10}
	mockRepo.On("FindAll", mock.Anything, expectedFilter).Return([]domain.Schedule{}, int64(0), nil)

	schedules, total, err := svc.ListSchedules(context.Background(), domain.ScheduleFilter{})

	assert.NoError(t, err)
	assert.Empty(t, schedules)
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}
