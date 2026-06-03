-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.subscriptions_log;;
-- +goose StatementEnd