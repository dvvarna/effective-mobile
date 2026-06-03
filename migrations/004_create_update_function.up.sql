-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.update_subscriptions(
    p_data json)
    RETURNS json
    LANGUAGE 'plpgsql'
    COST 100
    VOLATILE PARALLEL UNSAFE
AS $BODY$
DECLARE
    data_jsonb jsonb := p_data::jsonb;
    elem jsonb;
    res_array jsonb := '[]'::jsonb;
    v_id bigint;
    v_str text;
    set_clauses text[];
    sql_cmd text;
    cur_start_date date;
    cur_end_date date;
    err_code text;
    err_msg text;
    v_sqlstate text;
BEGIN
    FOR elem IN SELECT * FROM jsonb_array_elements(data_jsonb)
        LOOP
            BEGIN
                IF NOT elem ? 'user_subscription_id' THEN
                    res_array := res_array || (elem || jsonb_build_object(
                            'error_code', 'U0001',
                            'error_text', 'Отсутствует обязательное поле user_subscription_id'
                                                       ));
                    CONTINUE;
                END IF;

                v_id := (elem->>'user_subscription_id')::bigint;
                set_clauses := ARRAY[]::text[];

                -- service_name (NOT NULL)
                IF elem ? 'service_name' THEN
                    v_str := elem->>'service_name';
                    IF v_str = '' OR v_str = 'NULL' THEN
                        RAISE EXCEPTION 'service_name не может быть NULL' USING ERRCODE = 'U0002';
                    END IF;
                    set_clauses := array_append(set_clauses, format('service_name = %L', v_str));
                END IF;

                -- price (NOT NULL)
                IF elem ? 'price' THEN
                    v_str := elem->>'price';
                    IF v_str = '' OR v_str = 'NULL' THEN
                        RAISE EXCEPTION 'price не может быть NULL' USING ERRCODE = 'U0003';
                    END IF;
                    BEGIN
                        PERFORM v_str::int;
                    EXCEPTION WHEN OTHERS THEN
                        RAISE EXCEPTION 'Некорректное значение price' USING ERRCODE = 'U0004';
                    END;
                    set_clauses := array_append(set_clauses, format('price = %s', v_str::int));
                END IF;

                -- user_id (NOT NULL, UUID)
                IF elem ? 'user_id' THEN
                    v_str := elem->>'user_id';
                    IF v_str = '' OR v_str = 'NULL' THEN
                        RAISE EXCEPTION 'user_id не может быть NULL' USING ERRCODE = 'U0005';
                    END IF;
                    BEGIN
                        PERFORM v_str::uuid;
                    EXCEPTION WHEN OTHERS THEN
                        RAISE EXCEPTION 'Некорректный формат user_id' USING ERRCODE = 'U0006';
                    END;
                    set_clauses := array_append(set_clauses, format('user_id = %L', v_str::uuid));
                END IF;

                -- start_date (NOT NULL, приводим к первому числу месяца)
                IF elem ? 'start_date' THEN
                    v_str := elem->>'start_date';
                    IF v_str = '' OR v_str = 'NULL' THEN
                        RAISE EXCEPTION 'start_date не может быть NULL' USING ERRCODE = 'U0007';
                    END IF;
                    BEGIN
                        v_str := to_date(v_str, 'MM-YYYY')::text;
                        v_str := make_date(
                                extract(year from to_date(v_str, 'YYYY-MM-DD'))::int,
                                extract(month from to_date(v_str, 'YYYY-MM-DD'))::int,
                                1
                                 )::text;
                    EXCEPTION WHEN OTHERS THEN
                        RAISE EXCEPTION 'Некорректный формат start_date (ожидается MM-YYYY)' USING ERRCODE = 'U0008';
                    END;
                    set_clauses := array_append(set_clauses, format('start_date = %L', v_str::date));
                END IF;

                -- end_date (может быть NULL, приводим к последнему числу месяца или зачищаем)
                IF elem ? 'end_date' THEN
                    v_str := elem->>'end_date';
                    IF v_str = '' OR v_str = 'NULL' THEN
                        set_clauses := array_append(set_clauses, 'end_date = NULL');
                    ELSE
                        BEGIN
                            v_str := to_date(v_str, 'MM-YYYY')::text;
                            v_str := (
                                make_date(
                                        extract(year from to_date(v_str, 'YYYY-MM-DD'))::int,
                                        extract(month from to_date(v_str, 'YYYY-MM-DD'))::int,
                                        1
                                ) + interval '1 month' - interval '1 day'
                                )::date::text;
                        EXCEPTION WHEN OTHERS THEN
                            RAISE EXCEPTION 'Некорректный формат end_date (ожидается MM-YYYY)' USING ERRCODE = 'U0009';
                        END;
                        set_clauses := array_append(set_clauses, format('end_date = %L', v_str::date));
                    END IF;
                END IF;

                -- Проверка существования записи
                PERFORM 1 FROM public.subscriptions WHERE user_subscription_id = v_id;
                IF NOT FOUND THEN
                    RAISE EXCEPTION 'Запись с user_subscription_id=% не найдена', v_id USING ERRCODE = 'U0010';
                END IF;

                -- Выполнение UPDATE, если есть что менять
                IF array_length(set_clauses, 1) > 0 THEN
                    sql_cmd := 'UPDATE public.subscriptions SET '
                                   || array_to_string(set_clauses, ', ')
                                   || ' WHERE user_subscription_id = ' || v_id;
                    EXECUTE sql_cmd;

                    -- Бизнес-проверка: end_date >= start_date
                    SELECT start_date, end_date INTO cur_start_date, cur_end_date
                    FROM public.subscriptions WHERE user_subscription_id = v_id;
                    IF cur_end_date IS NOT NULL AND cur_end_date < cur_start_date THEN
                        RAISE EXCEPTION 'end_date (%) не может быть меньше start_date (%)',
                            cur_end_date, cur_start_date USING ERRCODE = 'U0011';
                    END IF;
                END IF;

                -- Успех
                res_array := res_array || (elem || '{"result":"OK"}'::jsonb);

            EXCEPTION WHEN OTHERS THEN
                GET STACKED DIAGNOSTICS
                    err_msg = MESSAGE_TEXT,
                    v_sqlstate = RETURNED_SQLSTATE;
                err_code := v_sqlstate;
                res_array := res_array || (elem || jsonb_build_object(
                        'error_code', err_code,
                        'error_text', err_msg
                                                   ));
            END;
        END LOOP;

    RETURN res_array::json;
END;
$BODY$;

ALTER FUNCTION public.update_subscriptions(json)
    OWNER TO postgres;
-- +goose StatementEnd