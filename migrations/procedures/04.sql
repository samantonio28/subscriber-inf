-- Хранимая процедура доступа к метаданным
-- Задача: Процедура, которая выводит список всех таблиц в текущей базе данных и количество строк в каждой из них.
CREATE OR REPLACE PROCEDURE get_tables_row_count()
LANGUAGE plpgsql
AS $$
DECLARE
    table_record RECORD;
    row_count BIGINT;
    sql_text TEXT;
BEGIN
    FOR table_record IN (
        SELECT tablename
        FROM pg_tables
        WHERE schemaname = 'public' -- Смотрим только таблицы в схеме public
        ORDER BY tablename
    )
    LOOP
        sql_text := 'SELECT COUNT(*) FROM ' || quote_ident(table_record.tablename);
        EXECUTE sql_text INTO row_count;
        RAISE NOTICE 'Таблица: %, Количество строк: %', table_record.tablename, row_count;
    END LOOP;
END;
$$;
