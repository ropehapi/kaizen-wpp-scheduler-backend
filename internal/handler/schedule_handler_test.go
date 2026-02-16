package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/internal/domain"
	"github.com/ropehapi/kaizen-wpp-scheduler-backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockScheduleService é o mock da camada de service para testes de handler.
type MockScheduleService struct {
	mock.Mock
}

func (m *MockScheduleService) CreateSchedule(ctx context.Context, req domain.CreateScheduleRequest) (*domain.Schedule, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Schedule), args.Error(1)
}

func (m *MockScheduleService) ListSchedules(ctx context.Context, filter domain.ScheduleFilter) ([]domain.Schedule, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domain.Schedule), args.Get(1).(int64), args.Error(2)
}

func (m *MockScheduleService) GetScheduleByID(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Schedule), args.Error(1)
}

func (m *MockScheduleService) UpdateSchedule(ctx context.Context, id uuid.UUID, req domain.UpdateScheduleRequest) (*domain.Schedule, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Schedule), args.Error(1)
}

func (m *MockScheduleService) CancelSchedule(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Schedule), args.Error(1)
}

// setupTestRouter cria um router de teste com o mock service injetado.
func setupTestRouter(mockSvc *MockScheduleService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewScheduleHandler(mockSvc)
	router := SetupRouter(h, "*")
	return router
}

// ===== Testes de CreateSchedule Handler =====

func TestCreateScheduleHandler_Success(t *testing.T) {
	mockSvc := new(MockScheduleService)
	router := setupTestRouter(mockSvc)

	scheduleID := uuid.New()
	body := map[string]interface{}{
		"message":     "Olá, mundo!",
		"type":        "once",
		"scheduledAt": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"contacts": []map[string]string{
			{"name": "João", "phone": "+5511999999999"},
		},
	}

	mockSvc.On("CreateSchedule", mock.Anything, mock.AnythingOfType("domain.CreateScheduleRequest")).
		Return(&domain.Schedule{
			ID:      scheduleID,
			Message: "Olá, mundo!",
			Type:    domain.ScheduleTypeOnce,
			Status:  domain.ScheduleStatusScheduled,
		}, nil)

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/schedules", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp["data"])
	assert.Nil(t, resp["error"])
}

func TestCreateScheduleHandler_ValidationError(t *testing.T) {
	mockSvc := new(MockScheduleService)
	router := setupTestRouter(mockSvc)

	// Body sem campo obrigatório "message"
	body := map[string]interface{}{
		"type":        "once",
		"scheduledAt": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		"contacts": []map[string]string{
			{"name": "João", "phone": "+5511999999999"},
		},
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/schedules", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Nil(t, resp["data"])
	assert.NotNil(t, resp["error"])
}

// ===== Testes de GetScheduleByID Handler =====

func TestGetScheduleByIDHandler_Success(t *testing.T) {
	mockSvc := new(MockScheduleService)
	router := setupTestRouter(mockSvc)

	scheduleID := uuid.New()
	mockSvc.On("GetScheduleByID", mock.Anything, scheduleID).
		Return(&domain.Schedule{
			ID:      scheduleID,
			Message: "Teste",
			Status:  domain.ScheduleStatusScheduled,
		}, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/schedules/"+scheduleID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetScheduleByIDHandler_InvalidUUID(t *testing.T) {
	mockSvc := new(MockScheduleService)
	router := setupTestRouter(mockSvc)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/schedules/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetScheduleByIDHandler_NotFound(t *testing.T) {
	mockSvc := new(MockScheduleService)
	router := setupTestRouter(mockSvc)

	scheduleID := uuid.New()
	mockSvc.On("GetScheduleByID", mock.Anything, scheduleID).
		Return(nil, service.ErrScheduleNotFound)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/schedules/"+scheduleID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ===== Testes de CancelSchedule Handler =====

func TestCancelScheduleHandler_Success(t *testing.T) {
	mockSvc := new(MockScheduleService)
	router := setupTestRouter(mockSvc)

	scheduleID := uuid.New()
	mockSvc.On("CancelSchedule", mock.Anything, scheduleID).
		Return(&domain.Schedule{
			ID:     scheduleID,
			Status: domain.ScheduleStatusCanceled,
		}, nil)

	req, _ := http.NewRequest(http.MethodPatch, "/api/v1/schedules/"+scheduleID.String()+"/cancel", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ===== Teste de Healthcheck =====

func TestHealthcheck(t *testing.T) {
	mockSvc := new(MockScheduleService)
	router := setupTestRouter(mockSvc)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "ok", data["status"])
}
