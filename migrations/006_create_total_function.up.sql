-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.get_total_subscriptions(
    p_data json)
    RETURNS json
    LANGUAGE 'plpgsql'
    COST 100
    STABLE PARALLEL UNSAFE
AS $BODY$
DECLARE
    start_date_str  text;
    end_date_str    text;
    start_date_val  date;
    end_date_val    date;
    user_id_val     uuid;
    service_name_val text;
    total_sum       bigint;
BEGIN
    -- 1. Извлекаем обязательные параметры
    start_date_str := p_data->>'start_date';
    end_date_str   := p_data->>'end_date';

    -- 2. Проверка наличия start_date и end_date
    IF start_date_str IS NULL OR end_date_str IS NULL THEN
        RETURN json_build_object(
                'start_date',   start_date_str,
                'end_date',     end_date_str,
                'user_id',      p_data->>'user_id',
                'service_name', p_data->>'service_name',
                'error_code',   'U0003',
                'error_text',   'Не указаны обязательные параметры start_date и/или end_date'
               );
    END IF;

    -- 3. Проверка формата MM-YYYY
    IF start_date_str !~ '^\d{2}-\d{4}$' OR end_date_str !~ '^\d{2}-\d{4}$' THEN
        RETURN json_build_object(
                'start_date',   start_date_str,
                'end_date',     end_date_str,
                'user_id',      p_data->>'user_id',
                'service_name', p_data->>'service_name',
                'error_code',   'U0003',
                'error_text',   'Неверный формат даты, ожидается MM-YYYY'
               );
    END IF;

    -- 4. Преобразование в даты (первое и последнее число месяца)
    BEGIN
        start_date_val := to_date(start_date_str, 'MM-YYYY');
        end_date_val   := (to_date(end_date_str, 'MM-YYYY') + interval '1 month' - interval '1 day')::date;
    EXCEPTION WHEN OTHERS THEN
        RETURN json_build_object(
                'start_date',   start_date_str,
                'end_date',     end_date_str,
                'user_id',      p_data->>'user_id',
                'service_name', p_data->>'service_name',
                'error_code',   'U0003',
                'error_text',   'Неверное значение месяца/года в дате'
               );
    END;

    -- 5. Проверка порядка дат
    IF end_date_val < start_date_val THEN
        RETURN json_build_object(
                'start_date',   start_date_str,
                'end_date',     end_date_str,
                'user_id',      p_data->>'user_id',
                'service_name', p_data->>'service_name',
                'error_code',   'U0002',
                'error_text',   'end_date должен быть больше или равен start_date'
               );
    END IF;

    -- 6. Извлечение необязательных параметров с проверкой UUID
    BEGIN
        user_id_val := NULLIF(p_data->>'user_id', '')::uuid;
    EXCEPTION WHEN OTHERS THEN
        RETURN json_build_object(
                'start_date',   start_date_str,
                'end_date',     end_date_str,
                'user_id',      p_data->>'user_id',
                'service_name', p_data->>'service_name',
                'error_code',   'U0004',
                'error_text',   'Неверный формат user_id'
               );
    END;

    service_name_val := NULLIF(p_data->>'service_name', '');

    -- 7. Подсчёт суммарной стоимости за пересекающиеся месяцы
    SELECT SUM(
                   (
                       (EXTRACT(YEAR FROM age(
                               LEAST(COALESCE(sub.end_date, end_date_val), end_date_val),
                               GREATEST(sub.start_date, start_date_val)
                                          )) * 12 +
                        EXTRACT(MONTH FROM age(
                                LEAST(COALESCE(sub.end_date, end_date_val), end_date_val),
                                GREATEST(sub.start_date, start_date_val)
                                           ))) + 1
                       ) * sub.price
           ) INTO total_sum
    FROM public.subscriptions sub
    WHERE sub.start_date <= end_date_val
      AND (sub.end_date >= start_date_val OR sub.end_date IS NULL)
      AND (user_id_val IS NULL OR sub.user_id = user_id_val)
      AND (service_name_val IS NULL OR sub.service_name = service_name_val);

    -- 8. Если данные отсутствуют
    IF total_sum IS NULL THEN
        RETURN json_build_object(
                'start_date',   start_date_str,
                'end_date',     end_date_str,
                'user_id',      p_data->>'user_id',
                'service_name', p_data->>'service_name',
                'error_code',   'U0001',
                'error_text',   'Данные отсутствуют'
               );
    END IF;

    -- 9. Успешное завершение
    RETURN json_build_object(
            'start_date',   start_date_str,
            'end_date',     end_date_str,
            'user_id',      CASE WHEN user_id_val IS NOT NULL THEN user_id_val::text ELSE NULL END,
            'service_name', service_name_val,
            'total',        total_sum
           );

EXCEPTION WHEN OTHERS THEN
    -- 10. Серверная ошибка
    RETURN json_build_object(
            'start_date',   start_date_str,
            'end_date',     end_date_str,
            'user_id',      p_data->>'user_id',
            'service_name', p_data->>'service_name',
            'error_code',   SQLSTATE,
            'error_text',   SQLERRM
           );
END;
$BODY$;

ALTER FUNCTION public.get_total_subscriptions(json)
    OWNER TO postgres;

-- +goose StatementEnd