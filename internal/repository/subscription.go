package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"effective-mobile/internal/models"
)

type SubscriptionRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewSubscriptionRepository(db *sql.DB, logger *slog.Logger) *SubscriptionRepository {
	return &SubscriptionRepository{db: db, logger: logger}
}

func (r *SubscriptionRepository) CreateBatch(ctx context.Context, payload []byte) ([]models.CreateSubscriptionResponseItem, error) {
	r.logger.Info("вызов процедуры insert_subscriptions_v2", slog.Int("payload_size", len(payload)))

	var resultJSON string
	query := "SELECT public.insert_subscriptions_v2($1::json)"

	if err := r.db.QueryRowContext(ctx, query, payload).Scan(&resultJSON); err != nil {
		r.logger.Error("ошибка выполнения процедуры", slog.String("error", err.Error()))
		return nil, fmt.Errorf("db procedure error: %w", err)
	}

	var responses []models.CreateSubscriptionResponseItem
	if err := json.Unmarshal([]byte(resultJSON), &responses); err != nil {
		r.logger.Error("не удалось распарсить ответ процедуры", slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid db response format: %w", err)
	}

	return responses, nil
}

// DeleteBatch -------------------------------------------------------------------------------------------------------------
// DeleteBatch вызывает хранимую процедуру удаления и возвращает массив результатов
func (r *SubscriptionRepository) DeleteBatch(ctx context.Context, payload []byte) ([]models.DeleteSubscriptionResponseItem, error) {
	r.logger.Info("вызов процедуры delete_subscriptions", slog.Int("ids_count", len(payload)))

	var resultJSON string
	query := "SELECT public.delete_subscriptions($1::json)"

	// Передаём JSON как параметр, сканируем результат в строку
	if err := r.db.QueryRowContext(ctx, query, payload).Scan(&resultJSON); err != nil {
		r.logger.Error("ошибка выполнения процедуры удаления", slog.String("error", err.Error()))
		return nil, fmt.Errorf("db delete procedure error: %w", err)
	}

	var responses []models.DeleteSubscriptionResponseItem
	if err := json.Unmarshal([]byte(resultJSON), &responses); err != nil {
		r.logger.Error("не удалось распарсить ответ процедуры удаления", slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid db delete response format: %w", err)
	}

	r.logger.Info("процедура удаления отработала", slog.Int("results_count", len(responses)))
	return responses, nil
}

// UpdateBatch -------------------------------------------------------------------------------------------------------------
// UpdateBatch вызывает хранимую процедуру обновления и возвращает массив результатов
func (r *SubscriptionRepository) UpdateBatch(ctx context.Context, payload []byte) ([]models.UpdateSubscriptionResponseItem, error) {
	r.logger.Info("вызов процедуры update_subscriptions", slog.Int("payload_size", len(payload)))

	var resultJSON string
	query := "SELECT public.update_subscriptions($1::json)"

	// Передаём JSON как параметр, сканируем результат в строку
	if err := r.db.QueryRowContext(ctx, query, payload).Scan(&resultJSON); err != nil {
		r.logger.Error("ошибка выполнения процедуры обновления", slog.String("error", err.Error()))
		return nil, fmt.Errorf("db update procedure error: %w", err)
	}

	var responses []models.UpdateSubscriptionResponseItem
	if err := json.Unmarshal([]byte(resultJSON), &responses); err != nil {
		r.logger.Error("не удалось распарсить ответ процедуры обновления", slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid db update response format: %w", err)
	}

	r.logger.Info("процедура обновления отработала", slog.Int("results_count", len(responses)))
	return responses, nil
}

// ReadOne -------------------------------------------------------------------------------------------------------------
// ReadOne вызывает хранимую процедуру чтения и возвращает результат
func (r *SubscriptionRepository) ReadOne(ctx context.Context, payload []byte) (models.ReadSubscriptionResponse, error) {
	r.logger.Info("вызов процедуры read_subscription", slog.String("payload", string(payload)))

	var resultJSON string

	query := "SELECT public.read_one_subscriptions($1::json)"

	if err := r.db.QueryRowContext(ctx, query, payload).Scan(&resultJSON); err != nil {
		r.logger.Error("ошибка выполнения процедуры чтения", slog.String("error", err.Error()))
		return models.ReadSubscriptionResponse{}, fmt.Errorf("db read procedure error: %w", err)
	}

	var response models.ReadSubscriptionResponse
	if err := json.Unmarshal([]byte(resultJSON), &response); err != nil {
		r.logger.Error("не удалось распарсить ответ процедуры чтения", slog.String("error", err.Error()))
		return models.ReadSubscriptionResponse{}, fmt.Errorf("invalid db read response format: %w", err)
	}

	r.logger.Info("процедура чтения отработала", slog.Int64("subscription_id", response.UserSubscriptionID))
	return response, nil
}

// Aggregate -------------------------------------------------------------------------------------------------------------
// Aggregate вызывает хранимую процедуру подсчёта и возвращает результат
func (r *SubscriptionRepository) Aggregate(ctx context.Context, payload []byte) (models.AggregateResponse, error) {
	r.logger.Info("вызов процедуры get_total_subscriptions", slog.Int("payload_size", len(payload)))

	var resultJSON string
	query := "SELECT public.get_total_subscriptions($1::json)"

	if err := r.db.QueryRowContext(ctx, query, payload).Scan(&resultJSON); err != nil {
		r.logger.Error("ошибка выполнения процедуры агрегации", slog.String("error", err.Error()))
		return models.AggregateResponse{}, fmt.Errorf("db aggregate procedure error: %w", err)
	}

	var response models.AggregateResponse
	if err := json.Unmarshal([]byte(resultJSON), &response); err != nil {
		r.logger.Error("не удалось распарсить ответ процедуры агрегации", slog.String("error", err.Error()))
		return models.AggregateResponse{}, fmt.Errorf("invalid db aggregate response format: %w", err)
	}

	r.logger.Info("процедура агрегации отработала",
		slog.String("period", fmt.Sprintf("%s - %s", response.StartDate, response.EndDate)))
	return response, nil
}

// ListSubscriptions -------------------------------------------------------------------------------------------------------------
// ListSubscriptions формирует безопасный запрос с динамическим WHERE и LIMIT/OFFSET
func (r *SubscriptionRepository) ListSubscriptions(ctx context.Context, params *models.ListQueryParams) (*models.ListResponse, error) {
	// Базовые запросы
	countQuery := `SELECT COUNT(*) FROM subscriptions WHERE 1=1`
	dataQuery := `
		SELECT user_subscription_id, user_id::text, service_name, price,
			   TO_CHAR(start_date, 'MM-YYYY'),
			   CASE WHEN end_date IS NOT NULL THEN TO_CHAR(end_date, 'MM-YYYY') ELSE NULL END
		FROM subscriptions WHERE 1=1`

	var filterArgs []interface{} // только для WHERE
	idx := 1

	// Динамическое добавление фильтров (одинаково для обоих запросов)
	if params.UserID != "" {
		clause := fmt.Sprintf(" AND user_id = $%d::uuid", idx)
		countQuery += clause
		dataQuery += clause
		filterArgs = append(filterArgs, params.UserID)
		idx++
	}
	if params.ServiceName != "" {
		clause := fmt.Sprintf(" AND service_name = $%d", idx)
		countQuery += clause
		dataQuery += clause
		filterArgs = append(filterArgs, params.ServiceName)
		idx++
	}

	// Пагинация ТОЛЬКО для dataQuery
	dataQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
	dataArgs := append(filterArgs, params.Limit, params.Offset) // фильтры + лимит/офсет

	r.logger.Info("выполнение запроса списка",
		slog.Int("filter_args", len(filterArgs)),
		slog.Int("data_args", len(dataArgs)))

	// 1. Подсчёт общего количества (используем ТОЛЬКО фильтры)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, filterArgs...).Scan(&total); err != nil {
		r.logger.Error("ошибка подсчёта total", slog.String("error", err.Error()))
		return nil, fmt.Errorf("count query error: %w", err)
	}

	// 2. Выборка данных (фильтры + пагинация)
	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		r.logger.Error("ошибка выборки данных", slog.String("error", err.Error()))
		return nil, fmt.Errorf("data query error: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			
		}
	}(rows)

	items := make([]models.SubscriptionItem, 0, params.Limit)
	for rows.Next() {
		var item models.SubscriptionItem
		if err := rows.Scan(&item.UserSubscriptionID, &item.UserID, &item.ServiceName, &item.Price, &item.StartDate, &item.EndDate); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	r.logger.Info("список подписок получен", slog.Int("count", len(items)), slog.Int("total", total))
	return &models.ListResponse{
		Items: items,
		Pagination: models.PaginationInfo{
			Page:  params.Page,
			Limit: params.Limit,
			Total: total,
		},
	}, nil
}
