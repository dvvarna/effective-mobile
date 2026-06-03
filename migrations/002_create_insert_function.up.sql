-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.insert_subscriptions_v2(
    p_data json)
    RETURNS json
    LANGUAGE 'plpgsql'
    COST 100
    VOLATILE SECURITY DEFINER PARALLEL UNSAFE
    SET search_path=public
AS $BODY$
DECLARE
    elem           json;
    v_user_id      uuid;
    v_service      text;
    v_price        int;
    v_start_str    text;
    v_end_str      text;
    v_start_date   date;
    v_end_date     date;
    v_inserted_id  bigint;
    v_result       json;
    v_result_arr   json[] := '{}';
    v_err_code     text;
    v_err_msg      text;
BEGIN
    FOR elem IN SELECT * FROM json_array_elements(p_data)
        LOOP
            BEGIN
                -- 1. Проверка user_id
                IF elem->>'user_id' IS NULL OR trim(elem->>'user_id') = '' THEN
                    RAISE EXCEPTION 'user_id is required and must not be empty' USING ERRCODE = 'U0001';
                END IF;
                BEGIN
                    v_user_id := (elem->>'user_id')::uuid;
                EXCEPTION WHEN OTHERS THEN
                    RAISE EXCEPTION 'user_id must be a valid UUID' USING ERRCODE = 'U0002';
                END;

                -- 2. Проверка service_name
                IF elem->>'service_name' IS NULL OR trim(elem->>'service_name') = '' THEN
                    RAISE EXCEPTION 'service_name is required and must not be empty' USING ERRCODE = 'U0003';
                END IF;
                v_service := trim(elem->>'service_name');

                -- 3. Проверка price
                IF elem->>'price' IS NULL THEN
                    RAISE EXCEPTION 'price is required' USING ERRCODE = 'U0004';
                END IF;
                BEGIN
                    v_price := (elem->>'price')::int;
                EXCEPTION WHEN OTHERS THEN
                    RAISE EXCEPTION 'price must be a valid integer' USING ERRCODE = 'U0005';
                END;
                IF v_price < 0 THEN
                    RAISE EXCEPTION 'price must not be negative' USING ERRCODE = 'U0006';
                END IF;

                -- 4. Проверка start_date
                v_start_str := elem->>'start_date';
                IF v_start_str IS NULL OR trim(v_start_str) = '' THEN
                    RAISE EXCEPTION 'start_date is required and must not be empty' USING ERRCODE = 'U0007';
                END IF;
                BEGIN
                    v_start_date := make_date(
                            (substring(v_start_str from '\d{4}$'))::int,
                            (substring(v_start_str from '^\d{2}'))::int,
                            1
                                    );
                EXCEPTION WHEN OTHERS THEN
                    RAISE EXCEPTION 'Invalid start_date format, expected MM-YYYY' USING ERRCODE = 'U0008';
                END;

                -- 5. Обработка необязательного end_date
                v_end_str := elem->>'end_date';
                IF v_end_str IS NOT NULL AND trim(v_end_str) <> '' THEN
                    BEGIN
                        v_end_date := make_date(
                                (substring(v_end_str from '\d{4}$'))::int,
                                (substring(v_end_str from '^\d{2}'))::int,
                                1
                                      );
                        v_end_date := (date_trunc('month', v_end_date) + INTERVAL '1 month' - INTERVAL '1 day')::date;
                    EXCEPTION WHEN OTHERS THEN
                        RAISE EXCEPTION 'Invalid end_date format, expected MM-YYYY' USING ERRCODE = 'U0009';
                    END;

                    -- 6. Проверка: end_date не может быть раньше start_date
                    IF v_end_date < v_start_date THEN
                        RAISE EXCEPTION 'end_date must not be earlier than start_date' USING ERRCODE = 'U0010';
                    END IF;
                ELSE
                    v_end_date := NULL;
                END IF;

                -- 7. Простая вставка без проверок пересечений
                INSERT INTO public.subscriptions
                (user_id, service_name, price, start_date, end_date)
                VALUES
                    (v_user_id, v_service, v_price, v_start_date, v_end_date)
                RETURNING user_subscription_id INTO v_inserted_id;

                -- 8. Успешный ответ
                v_result := json_build_object(
                        'service_name', v_service,
                        'price', v_price,
                        'user_id', v_user_id,
                        'start_date', v_start_str,
                        'end_date', CASE WHEN v_end_str IS NOT NULL AND trim(v_end_str) <> '' THEN v_end_str ELSE NULL END,
                        'user_subscription_id', v_inserted_id
                            );
                v_result := json_strip_nulls(v_result);
                v_result_arr := v_result_arr || ARRAY[v_result];

            EXCEPTION WHEN OTHERS THEN
                -- Формируем ответ с ошибкой, сохраняя исходные значения без приведения типов
                v_err_code := SQLSTATE;
                v_err_msg  := SQLERRM;
                v_result := json_build_object(
                        'service_name', elem->>'service_name',
                        'price', to_jsonb(elem->'price'),
                        'user_id', elem->>'user_id',
                        'start_date', elem->>'start_date',
                        'end_date', CASE WHEN elem->>'end_date' IS NOT NULL AND trim(elem->>'end_date') <> '' THEN elem->>'end_date' ELSE NULL END,
                        'error_code', v_err_code,
                        'error_text', v_err_msg
                            );
                v_result := json_strip_nulls(v_result);
                v_result_arr := v_result_arr || ARRAY[v_result];
            END;
        END LOOP;

    RETURN array_to_json(v_result_arr)::json;
END;
$BODY$;

ALTER FUNCTION public.insert_subscriptions_v2(json)
    OWNER TO postgres;

COMMENT ON FUNCTION public.insert_subscriptions_v2(json)
    IS 'Пакетная вставка подписок (версия 2): строгая валидация полей (user_id, service_name, price, start_date, end_date>=start_date), без проверки пересечений.';
-- +goose StatementEnd