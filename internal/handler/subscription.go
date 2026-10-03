package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"effective-mobile/internal/service"

	"github.com/go-chi/chi/v5"
)

type SubscriptionHandler struct {
	svc    *service.SubscriptionService
	logger *slog.Logger
}

func NewSubscriptionHandler(svc *service.SubscriptionService, logger *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc, logger: logger}
}

// CreateSubscriptions godoc
// @Summary Create Subscriptions
// @Description Create Subscriptions
// @Tags Subscriptions
// @Accept JSON
// @Produce JSON
// @Param body []effective-mobile/internal/models.CreateSubscriptionRequestItem true "Array subscriptions"
// @Success 200 {array} effective-mobile/internal/models.CreateSubscriptionResponseItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscriptions(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // макс 1MB
	if err != nil {
		h.logger.Error("чтение тела запроса", slog.String("error", err.Error()))
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to read body"})
		return
	}

	ctx := r.Context()
	results, err := h.svc.CreateSubscriptions(ctx, body)
	if err != nil {
		h.logger.Error("ошибка сервиса", slog.String("error", err.Error()))

		status := http.StatusInternalServerError
		if err.Error() == "invalid request body: must be a JSON array" || err.Error() == "request body must not be empty" {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ListSubscriptions godoc
// @Summary List Subscriptions
// @Description Return List Subscriptions
// @Tags Subscriptions
// @Produce JSON
// @Param page query int false "Number page" default(1)
// @Param limit query int false "Limit rows" default(20)
// @Param user_id query string false "User UUID"
// @Param service_name query string false "Service Name"
// @Success 200 {object} effective-mobile/internal/models.ListResponse
// @Failure 500 {object} map[string]string
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	userID := q.Get("user_id")
	serviceName := q.Get("service_name")

	result, err := h.svc.ListSubscriptions(r.Context(), page, limit, userID, serviceName)
	if err != nil {
		h.logger.Error("ошибка получения списка", slog.String("error", err.Error()))

		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid user_id format") {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	h.logger.Info("список успешно сформирован", slog.Int("items", len(result.Items)))
	writeJSON(w, http.StatusOK, result)
}

// GetSubscription godoc
// @Summary Получить одну подписку
// @Description Возвращает данные подписки по ID
// @Tags Subscriptions
// @Produce JSON
// @Param user_subscription_id path int true "ID подписки"
// @Success 200 {object} effective-mobile/internal/models.ReadSubscriptionResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} effective-mobile/internal/models.ReadSubscriptionResponse
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{user_subscription_id} [get]
func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	// 1. Извлекаем ID из URL
	idStr := chi.URLParam(r, "user_subscription_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		h.logger.Warn("некорректный ID в URL", slog.String("id", idStr))
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or missing user_subscription_id"})
		return
	}

	// 2. Формируем JSON-пейлоад, как ожидает процедура
	payload, _ := json.Marshal(map[string]int64{"user_subscription_id": id})

	ctx := r.Context()
	result, err := h.svc.ReadOne(ctx, payload)
	if err != nil {
		h.logger.Error("ошибка сервиса чтения", slog.String("error", err.Error()))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// 3. Маппим error_code из БД на HTTP-статусы
	if result.ErrorCode != "" {
		status := http.StatusInternalServerError
		if result.ErrorCode == "U0001" {
			status = http.StatusNotFound // "Данные отсутствуют"
		}
		// Логируем ошибку на уровне HTTP
		h.logger.Warn("процедура вернула ошибку",
			slog.Int64("id", result.UserSubscriptionID),
			slog.String("code", result.ErrorCode),
			slog.String("text", result.ErrorText))
		writeJSON(w, status, result)
		return
	}

	// 4. Успех
	h.logger.Info("подписка успешно найдена", slog.Int64("id", result.UserSubscriptionID))
	writeJSON(w, http.StatusOK, result)
}

// UpdateSubscriptions godoc
// @Summary Обновление подписок
// @Description Частичное обновление одной или нескольких записей
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param body body []effective-mobile/internal/models.UpdateSubscriptionRequestItem true "Массив обновляемых подписок"
// @Success 200 {array} effective-mobile/internal/models.UpdateSubscriptionResponseItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [put]
func (h *SubscriptionHandler) UpdateSubscriptions(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	// Читаем тело запроса
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // макс 1MB
	if err != nil {
		h.logger.Error("чтение тела запроса на обновление", slog.String("error", err.Error()))
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to read body"})
		return
	}

	ctx := r.Context()
	results, err := h.svc.UpdateSubscriptions(ctx, body)
	if err != nil {
		h.logger.Error("ошибка сервиса обновления", slog.String("error", err.Error()))

		status := http.StatusInternalServerError
		if err.Error() == "invalid request body: must be a JSON array" ||
			err.Error() == "request body must not be empty" ||
			strings.HasPrefix(err.Error(), "item at index") {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	// Всегда 200, т.к. частичные ошибки обновления возвращаются в теле ответа
	// Аналогично POST/DELETE: успех/ошибка по каждому ID внутри массива
	writeJSON(w, http.StatusOK, results)
}

// DeleteSubscriptions godoc
// @Summary Удаление подписок
// @Description Удаляет одну или несколько записей по ID
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param body body effective-mobile/internal/models.DeleteSubscriptionRequest true "Массив ID для удаления"
// @Success 200 {array} effective-mobile/internal/models.DeleteSubscriptionResponseItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [delete]
func (h *SubscriptionHandler) DeleteSubscriptions(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	// Читаем тело запроса
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // макс 1MB
	if err != nil {
		h.logger.Error("чтение тела запроса на удаление", slog.String("error", err.Error()))
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to read body"})
		return
	}

	ctx := r.Context()
	results, err := h.svc.DeleteSubscriptions(ctx, body)
	if err != nil {
		h.logger.Error("ошибка сервиса удаления", slog.String("error", err.Error()))

		status := http.StatusInternalServerError
		if err.Error() == "invalid request body: must be JSON with user_subscription_id array" ||
			err.Error() == "request body must contain at least one user_subscription_id" {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	// Всегда 200, т.к. частичные ошибки удаления возвращаются в теле ответа
	// Аналогично POST: успех/ошибка по каждому ID внутри массива
	writeJSON(w, http.StatusOK, results)
}

// AggregateSubscriptions godoc
// @Summary Агрегация стоимости
// @Description Подсчёт суммарной стоимости подписок за период
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param body body effective-mobile/internal/models.AggregateRequest true "Период и фильтры"
// @Success 200 {object} effective-mobile/internal/models.AggregateResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/total [post]
func (h *SubscriptionHandler) AggregateSubscriptions(w http.ResponseWriter, r *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // макс 1MB
	if err != nil {
		h.logger.Error("чтение тела запроса на агрегацию", slog.String("error", err.Error()))
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to read body"})
		return
	}

	ctx := r.Context()
	result, err := h.svc.Aggregate(ctx, body)
	if err != nil {
		h.logger.Error("ошибка сервиса агрегации", slog.String("error", err.Error()))

		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid request body") ||
			strings.Contains(err.Error(), "start_date and end_date are required") {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	// Маппинг error_code из БД на HTTP-статусы
	if result.ErrorCode != "" {
		status := http.StatusInternalServerError
		if result.ErrorCode == "U0001" {
			status = http.StatusBadRequest // Пользовательская ошибка валидации/фильтров
		}
		h.logger.Warn("процедура агрегации вернула ошибку",
			slog.String("code", result.ErrorCode),
			slog.String("text", result.ErrorText))
		writeJSON(w, status, result)
		return
	}

	// Успех
	h.logger.Info("агрегация успешно выполнена", slog.Any("total", result.Total))
	writeJSON(w, http.StatusOK, result)
}
