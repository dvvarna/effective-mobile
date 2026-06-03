-- +goose Up
-- +goose StatementBegin

-- Таблица
CREATE TABLE IF NOT EXISTS public.subscriptions_log
(
    user_subscription_id bigint NOT NULL,
    user_id uuid NOT NULL,
    service_name character varying(100) COLLATE pg_catalog."default" NOT NULL,
    price integer NOT NULL,
    start_date date NOT NULL,
    end_date date,
    create_date timestamp with time zone NOT NULL,
    operation character(1) COLLATE pg_catalog."default" NOT NULL,
    operation_date timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    operation_user character varying(100) COLLATE pg_catalog."default" NOT NULL DEFAULT CURRENT_USER,
    CONSTRAINT "PK_subscriptions_log" PRIMARY KEY (user_subscription_id, operation, operation_date, operation_user)
)

    TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.subscriptions_log
    OWNER to postgres;

COMMENT ON TABLE public.subscriptions_log
    IS 'Подписки пользователей - лог изменений';
-- +goose StatementEnd