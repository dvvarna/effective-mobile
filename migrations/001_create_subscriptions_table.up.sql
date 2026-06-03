-- +goose Up
-- +goose StatementBegin

-- Таблица
CREATE TABLE IF NOT EXISTS public.subscriptions (
    user_subscription_id BIGSERIAL PRIMARY KEY,
    user_id uuid NOT NULL,
    service_name VARCHAR(100) NOT NULL,
    price integer NOT NULL  CHECK (price >= 0),
    start_date date NOT NULL,
    end_date date,
    create_date timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_price_positive CHECK (price >= 0)
);

ALTER TABLE IF EXISTS public.subscriptions
    OWNER to postgres;

COMMENT ON TABLE public.subscriptions
    IS 'Подписки пользователей';


-- Индексы
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON public.subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON public.subscriptions(service_name);
CREATE INDEX IF NOT EXISTS idx_subscriptions_date_range ON public.subscriptions(start_date, end_date);
-- +goose StatementEnd