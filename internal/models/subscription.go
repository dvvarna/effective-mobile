package models

type CreateSubscriptionRequestItem struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

type CreateSubscriptionResponseItem struct {
	ServiceName        string  `json:"service_name"`
	Price              int     `json:"price"`
	UserID             string  `json:"user_id"`
	StartDate          string  `json:"start_date"`
	EndDate            *string `json:"end_date,omitempty"`
	UserSubscriptionID *int64  `json:"user_subscription_id,omitempty"`
	ErrorCode          *int    `json:"error_code,omitempty"`
	ErrorText          *string `json:"error_text,omitempty"`
}

// DeleteSubscriptionRequest -------------------------------------------------------------------
// DeleteSubscriptionRequest запрос на удаление (массив ID)
type DeleteSubscriptionRequest struct {
	UserSubscriptionIDs []int64 `json:"user_subscription_id"`
}

// DeleteSubscriptionResponseItem элемент ответа для одного ID
type DeleteSubscriptionResponseItem struct {
	UserSubscriptionID int64  `json:"user_subscription_id"`
	Result             string `json:"result,omitempty"`
	ErrorCode          string `json:"error_code,omitempty"`
	ErrorText          string `json:"error_text,omitempty"`
}

// UpdateSubscriptionRequestItem -------------------------------------------------------------------
// UpdateSubscriptionRequestItem элемент массива на вход (обновление)
// user_subscription_id обязателен, остальные поля опциональны
type UpdateSubscriptionRequestItem struct {
	UserSubscriptionID int64   `json:"user_subscription_id"`
	ServiceName        *string `json:"service_name,omitempty"`
	Price              *int    `json:"price,omitempty"`
	UserID             *string `json:"user_id,omitempty"`
	StartDate          *string `json:"start_date,omitempty"`
	EndDate            *string `json:"end_date,omitempty"`
}

// UpdateSubscriptionResponseItem элемент массива на выход
type UpdateSubscriptionResponseItem struct {
	UserSubscriptionID int64   `json:"user_subscription_id"`
	ServiceName        string  `json:"service_name,omitempty"`
	Price              int     `json:"price,omitempty"`
	UserID             string  `json:"user_id,omitempty"`
	StartDate          string  `json:"start_date,omitempty"`
	EndDate            *string `json:"end_date,omitempty"`
	Result             string  `json:"result,omitempty"`
	ErrorCode          string  `json:"error_code,omitempty"`
	ErrorText          string  `json:"error_text,omitempty"`
}

// ReadSubscriptionRequest -------------------------------------------------------------------
// ReadSubscriptionRequest запрос на чтение (ID обязателен)
type ReadSubscriptionRequest struct {
	UserSubscriptionID int64 `json:"user_subscription_id"`
}

// ReadSubscriptionResponse ответ процедуры (данные или ошибка)
type ReadSubscriptionResponse struct {
	UserSubscriptionID int64   `json:"user_subscription_id"`
	UserID             string  `json:"user_id,omitempty"`
	ServiceName        string  `json:"service_name,omitempty"`
	Price              int     `json:"price,omitempty"`
	StartDate          string  `json:"start_date,omitempty"`
	EndDate            *string `json:"end_date,omitempty"`
	ErrorCode          string  `json:"error_code,omitempty"`
	ErrorText          string  `json:"error_text,omitempty"`
}

// AggregateRequest -------------------------------------------------------------------
// AggregateRequest запрос на подсчёт суммы
type AggregateRequest struct {
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	UserID      *string `json:"user_id,omitempty"`
	ServiceName *string `json:"service_name,omitempty"`
}

// AggregateResponse ответ процедуры (успех или ошибка)
type AggregateResponse struct {
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	UserID      *string `json:"user_id,omitempty"`
	ServiceName *string `json:"service_name,omitempty"`
	Total       *int    `json:"total,omitempty"`
	ErrorCode   string  `json:"error_code,omitempty"`
	ErrorText   string  `json:"error_text,omitempty"`
}

// ListQueryParams -------------------------------------------------------------------
// ListQueryParams параметры запроса списка
type ListQueryParams struct {
	Page        int
	Limit       int
	Offset      int
	UserID      string
	ServiceName string
}

// SubscriptionItem элемент списка
type SubscriptionItem struct {
	UserSubscriptionID int64   `json:"user_subscription_id"`
	UserID             string  `json:"user_id"`
	ServiceName        string  `json:"service_name"`
	Price              int     `json:"price"`
	StartDate          string  `json:"start_date"`
	EndDate            *string `json:"end_date,omitempty"`
}

// PaginationInfo мета-информация о пагинации
type PaginationInfo struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// ListResponse обёртка ответа списка
type ListResponse struct {
	Items      []SubscriptionItem `json:"items"`
	Pagination PaginationInfo     `json:"pagination"`
}
