package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"effective-mobile/internal/models"
	"effective-mobile/internal/repository"

	"github.com/google/uuid"
)

type SubscriptionService struct {
	repo   *repository.SubscriptionRepository
	logger *slog.Logger
}

func NewSubscriptionService(repo *repository.SubscriptionRepository, logger *slog.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, logger: logger}
}

func (s *SubscriptionService) CreateSubscriptions(ctx context.Context, body []byte) ([]models.CreateSubscriptionResponseItem, error) {
	var items []models.CreateSubscriptionRequestItem
	if err := json.Unmarshal(body, &items); err != nil {
		s.logger.Warn("получен некорректный JSON", slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid request body: must be a JSON array")
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("request body must not be empty")
	}

	s.logger.Info("передача пакета в БД", slog.Int("count", len(items)))
	return s.repo.CreateBatch(ctx, body)
}

// DeleteSubscriptions --------------------------------------------------------------------------------------
// DeleteSubscriptions валидирует запрос и делегирует удаление в БД
func (s *SubscriptionService) DeleteSubscriptions(ctx context.Context, body []byte) ([]models.DeleteSubscriptionResponseItem, error) {
	// Проверяем, что это валидный JSON с массивом ID
	var req models.DeleteSubscriptionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.logger.Warn("получен некорректный JSON для удаления", slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid request body: must be JSON with user_subscription_id array")
	}

	if len(req.UserSubscriptionIDs) == 0 {
		return nil, fmt.Errorf("request body must contain at least one user_subscription_id")
	}

	s.logger.Info("передача пакета на удаление в БД", slog.Int("count", len(req.UserSubscriptionIDs)))
	return s.repo.DeleteBatch(ctx, body)
}

// UpdateSubscriptions --------------------------------------------------------------------------------------
// UpdateSubscriptions валидирует запрос и делегирует обновление в БД
func (s *SubscriptionService) UpdateSubscriptions(ctx context.Context, body []byte) ([]models.UpdateSubscriptionResponseItem, error) {
	// Проверяем, что это валидный JSON-массив
	var items []models.UpdateSubscriptionRequestItem
	if err := json.Unmarshal(body, &items); err != nil {
		s.logger.Warn("получен некорректный JSON для обновления", slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid request body: must be a JSON array")
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("request body must not be empty")
	}

	// Валидация: каждый элемент должен иметь user_subscription_id
	for i, item := range items {
		if item.UserSubscriptionID <= 0 {
			s.logger.Warn("некорректный user_subscription_id", slog.Int("index", i), slog.Int64("id", item.UserSubscriptionID))
			return nil, fmt.Errorf("item at index %d: user_subscription_id must be positive", i)
		}
	}

	s.logger.Info("передача пакета на обновление в БД", slog.Int("count", len(items)))
	return s.repo.UpdateBatch(ctx, body)
}

// ReadOne --------------------------------------------------------------------------------------
// ReadOne валидирует запрос и делегирует чтение в БД
func (s *SubscriptionService) ReadOne(ctx context.Context, body []byte) (models.ReadSubscriptionResponse, error) {
	var req models.ReadSubscriptionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.logger.Warn("получен некорректный JSON для чтения", slog.String("error", err.Error()))
		return models.ReadSubscriptionResponse{}, fmt.Errorf("invalid request body: must contain user_subscription_id")
	}

	if req.UserSubscriptionID <= 0 {
		return models.ReadSubscriptionResponse{}, fmt.Errorf("user_subscription_id must be positive")
	}

	s.logger.Info("передача запроса на чтение в БД", slog.Int64("id", req.UserSubscriptionID))
	return s.repo.ReadOne(ctx, body)
}

// Aggregate --------------------------------------------------------------------------------------
// Aggregate валидирует обязательные поля и делегирует расчёт в БД
func (s *SubscriptionService) Aggregate(ctx context.Context, body []byte) (models.AggregateResponse, error) {
	var req models.AggregateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.logger.Warn("получен некорректный JSON для агрегации", slog.String("error", err.Error()))
		return models.AggregateResponse{}, fmt.Errorf("invalid request body")
	}

	// start_date и end_date обязательны
	if req.StartDate == "" || req.EndDate == "" {
		return models.AggregateResponse{}, fmt.Errorf("start_date and end_date are required")
	}

	s.logger.Info("передача запроса на агрегацию в БД",
		slog.String("start_date", req.StartDate),
		slog.String("end_date", req.EndDate))
	return s.repo.Aggregate(ctx, body)
}

// ListSubscriptions --------------------------------------------------------------------
const MaxListLimit = 100 // Жёсткий лимит для защиты БД

// ListSubscriptions валидирует параметры, рассчитывает offset и делегирует в БД
func (s *SubscriptionService) ListSubscriptions(ctx context.Context, rawPage, rawLimit int, userID, serviceName string) (*models.ListResponse, error) {
	page := rawPage
	limit := rawLimit

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > MaxListLimit {
		limit = MaxListLimit
	}
	offset := (page - 1) * limit

	// Базовая валидация UUID (если передан)
	if userID != "" {
		if _, err := uuid.Parse(userID); err != nil {
			return nil, fmt.Errorf("invalid user_id format")
		}
	}

	params := &models.ListQueryParams{
		Page:        page,
		Limit:       limit,
		Offset:      offset,
		UserID:      userID,
		ServiceName: serviceName,
	}

	s.logger.Info("запрос списка подписок",
		slog.Int("page", page), slog.Int("limit", limit),
		slog.String("user_id", userID), slog.String("service_name", serviceName))

	return s.repo.ListSubscriptions(ctx, params)
}
