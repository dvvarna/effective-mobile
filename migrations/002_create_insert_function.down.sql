-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS public.insert_subscriptions_v2(JSON);
-- +goose StatementEnd