-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION log_subscription_changes()
    RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        INSERT INTO public.subscriptions_log (
            user_subscription_id,
            user_id,
            service_name,
            price,
            start_date,
            end_date,
            create_date,
            operation
        ) VALUES (
                     OLD.user_subscription_id,
                     OLD.user_id,
                     OLD.service_name,
                     OLD.price,
                     OLD.start_date,
                     OLD.end_date,
                     OLD.create_date,
                     'U'
                 );
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO public.subscriptions_log (
            user_subscription_id,
            user_id,
            service_name,
            price,
            start_date,
            end_date,
            create_date,
            operation
        ) VALUES (
                     OLD.user_subscription_id,
                     OLD.user_id,
                     OLD.service_name,
                     OLD.price,
                     OLD.start_date,
                     OLD.end_date,
                     OLD.create_date,
                     'D'
                 );
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Триггер на изменение (старая версия строки попадает в лог)
CREATE TRIGGER trg_subscriptions_update
    AFTER UPDATE ON public.subscriptions
    FOR EACH ROW
EXECUTE FUNCTION log_subscription_changes();

-- Триггер на удаление (удалённая строка попадает в лог)
CREATE TRIGGER trg_subscriptions_delete
    AFTER DELETE ON public.subscriptions
    FOR EACH ROW
EXECUTE FUNCTION log_subscription_changes();

-- +goose StatementEnd