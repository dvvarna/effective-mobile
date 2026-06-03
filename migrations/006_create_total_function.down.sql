-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS public.get_total_subscriptions(JSON);
-- +goose StatementEnd