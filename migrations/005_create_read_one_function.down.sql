-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS public.read_one_subscriptions(JSON);
-- +goose StatementEnd