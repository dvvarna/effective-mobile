-- +goose Down
-- +goose StatementBegin
--DROP TRIGGER  IF EXISTS trg_subscriptions_update();
--DROP TRIGGER  IF EXISTS trg_subscriptions_delete();
DROP FUNCTION IF EXISTS public.log_subscription_changes();
-- +goose StatementEnd