-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS public.update_subscriptions(JSON);
-- +goose StatementEnd