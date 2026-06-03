-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS public.delete_subscriptions(JSON);
-- +goose StatementEnd