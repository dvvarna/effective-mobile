-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.read_one_subscriptions(
    p_data json)
    RETURNS json
    LANGUAGE 'plpgsql'
    COST 100
    STABLE PARALLEL UNSAFE
AS $BODY$
DECLARE
    v_sub_id bigint;
    v_row record;
    v_result json;
BEGIN
    -- Извлекаем идентификатор подписки из входного JSON
    v_sub_id := (p_data ->> 'user_subscription_id')::bigint;

    -- Пытаемся найти запись в таблице
    SELECT user_subscription_id,
           user_id,
           service_name,
           price,
           start_date,
           end_date
    INTO v_row
    FROM public.subscriptions
    WHERE user_subscription_id = v_sub_id;

    -- Если запись не найдена, возвращаем пользовательскую ошибку
    IF NOT FOUND THEN
        v_result := json_build_object(
                'user_subscription_id', v_sub_id,
                'error_code', 'U0001',
                'error_text', 'Данные отсутствуют'
                    );
        RETURN v_result;
    END IF;

    -- Формируем успешный ответ с преобразованием дат в формат MM-YYYY
    v_result := json_build_object(
            'user_subscription_id', v_row.user_subscription_id,
            'user_id', v_row.user_id,
            'service_name', v_row.service_name,
            'price', v_row.price,
            'start_date', to_char(v_row.start_date, 'MM-YYYY'),
            'end_date', CASE
                            WHEN v_row.end_date IS NOT NULL
                                THEN to_char(v_row.end_date, 'MM-YYYY')
                            ELSE NULL
                END
                );

    RETURN v_result;

EXCEPTION
    WHEN OTHERS THEN
        -- При любой другой ошибке возвращаем реальную ошибку сервера
        v_result := json_build_object(
                'user_subscription_id', v_sub_id,
                'error_code', SQLSTATE,
                'error_text', SQLERRM
                    );
        RETURN v_result;
END;
$BODY$;

ALTER FUNCTION public.read_one_subscriptions(json)
    OWNER TO postgres;

COMMENT ON FUNCTION public.read_one_subscriptions(json)
    IS 'Чтение одной записи подписки по user_subscription_id. Принимает и возвращает JSON.';
-- +goose StatementEnd