-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_subscriptions_date_range;
DROP INDEX IF EXISTS idx_subscriptions_service_name;
DROP INDEX IF EXISTS idx_subscriptions_user_id;
DROP TABLE IF EXISTS public.subscriptions;
-- +goose StatementEnd