-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.delete_subscriptions(p_json json)
    RETURNS json
    LANGUAGE plpgsql
AS $$
DECLARE
    ids json;
    id_elem json;
    v_id bigint;
    v_result_text text;
    v_results_text text := '[]';
    v_is_first boolean := true;
    v_error_text text;
    v_error_code text;
BEGIN
    -- Извлекаем массив идентификаторов
    ids := p_json->'user_subscription_id';
    IF ids IS NULL OR json_typeof(ids) != 'array' THEN
        RAISE EXCEPTION 'Invalid input: expected JSON with key "user_subscription_id" as array';
    END IF;

    -- Проверяем, что все элементы массива — числа
    FOR id_elem IN SELECT * FROM json_array_elements(ids)
        LOOP
            IF json_typeof(id_elem) != 'number' THEN
                RAISE EXCEPTION 'Invalid element in array: expected numbers, got %', json_typeof(id_elem);
            END IF;

            v_id := id_elem::text::bigint; -- Безопасно, так как элемент — число
            BEGIN
                DELETE FROM public.subscriptions WHERE user_subscription_id = v_id;
                v_result_text := json_build_object('user_subscription_id', v_id, 'result', 'OK')::text;
            EXCEPTION
                WHEN OTHERS THEN
                    GET STACKED DIAGNOSTICS
                        v_error_text = MESSAGE_TEXT,
                        v_error_code = RETURNED_SQLSTATE;
                    v_result_text := json_build_object(
                            'user_subscription_id', v_id,
                            'error_code', v_error_code,
                            'error_text', v_error_text
                                     )::text;
            END;

            -- Собираем итоговый JSON-массив в виде текста
            IF v_is_first THEN
                v_results_text := '[' || v_result_text;
                v_is_first := false;
            ELSE
                v_results_text := v_results_text || ',' || v_result_text;
            END IF;
        END LOOP;

    v_results_text := v_results_text || ']';
    RETURN v_results_text::json;
END;
$$;
-- +goose StatementEnd