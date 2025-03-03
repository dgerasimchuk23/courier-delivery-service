@echo off
REM Скрипт для анализа производительности базы данных (Windows)

REM Параметры подключения к базе данных
set DB_HOST=localhost
set DB_PORT=5432
set DB_NAME=delivery
set DB_USER=postgres
set DB_PASSWORD=postgres

REM Путь к файлу с SQL-запросами
set SQL_FILE=..\internal\db\analyze_queries.sql

REM Проверка наличия файла с SQL-запросами
if not exist "%SQL_FILE%" (
    echo Ошибка: Файл %SQL_FILE% не найден
    exit /b 1
)

REM Создание директории для результатов
set TIMESTAMP=%date:~-4,4%%date:~-7,2%%date:~-10,2%_%time:~0,2%%time:~3,2%%time:~6,2%
set TIMESTAMP=%TIMESTAMP: =0%
set RESULTS_DIR=db_analysis_%TIMESTAMP%
mkdir %RESULTS_DIR%
echo Результаты будут сохранены в директории: %RESULTS_DIR%

REM Анализ индексов
echo Выполнение запроса: Анализ индексов
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "SELECT schemaname, tablename, indexname, indexdef FROM pg_indexes WHERE schemaname = 'public' ORDER BY tablename, indexname;" > "%RESULTS_DIR%\indexes.txt"
echo Результат сохранен в файл: %RESULTS_DIR%\indexes.txt

REM Анализ размера таблиц
echo Выполнение запроса: Анализ размера таблиц
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "SELECT relname AS table_name, pg_size_pretty(pg_total_relation_size(relid)) AS total_size, pg_size_pretty(pg_relation_size(relid)) AS table_size, pg_size_pretty(pg_total_relation_size(relid) - pg_relation_size(relid)) AS index_size FROM pg_catalog.pg_statio_user_tables ORDER BY pg_total_relation_size(relid) DESC;" > "%RESULTS_DIR%\table_sizes.txt"
echo Результат сохранен в файл: %RESULTS_DIR%\table_sizes.txt

REM Анализ запросов для таблицы users
echo Выполнение запроса: Анализ запросов для таблицы users
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';" > "%RESULTS_DIR%\query_users_email.txt"
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "EXPLAIN ANALYZE SELECT * FROM users WHERE id = 1;" > "%RESULTS_DIR%\query_users_id.txt"
echo Результаты сохранены в файлы: %RESULTS_DIR%\query_users_*.txt

REM Анализ запросов для таблицы refresh_tokens
echo Выполнение запроса: Анализ запросов для таблицы refresh_tokens
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "EXPLAIN ANALYZE SELECT * FROM refresh_tokens WHERE token = 'test-token';" > "%RESULTS_DIR%\query_refresh_tokens_token.txt"
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "EXPLAIN ANALYZE SELECT * FROM refresh_tokens WHERE user_id = 1;" > "%RESULTS_DIR%\query_refresh_tokens_user_id.txt"
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "EXPLAIN ANALYZE SELECT * FROM refresh_tokens WHERE expires_at < NOW();" > "%RESULTS_DIR%\query_refresh_tokens_expires.txt"
echo Результаты сохранены в файлы: %RESULTS_DIR%\query_refresh_tokens_*.txt

REM Анализ сложных запросов с JOIN
echo Выполнение запроса: Анализ сложных запросов с JOIN
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "EXPLAIN ANALYZE SELECT d.*, p.address, c.name AS courier_name, cu.name AS customer_name FROM delivery d JOIN parcel p ON d.parcel_id = p.id JOIN courier c ON d.courier_id = c.id JOIN customer cu ON p.client = cu.id WHERE d.status = 'delivered';" > "%RESULTS_DIR%\query_complex_join.txt"
echo Результат сохранен в файл: %RESULTS_DIR%\query_complex_join.txt

REM Проверка блокировок
echo Выполнение запроса: Проверка блокировок
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "SELECT blocked_locks.pid AS blocked_pid, blocking_locks.pid AS blocking_pid, blocked_activity.usename AS blocked_user, blocking_activity.usename AS blocking_user, blocked_activity.query AS blocked_statement, blocking_activity.query AS blocking_statement FROM pg_catalog.pg_locks blocked_locks JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid AND blocking_locks.pid != blocked_locks.pid JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid WHERE NOT blocked_locks.granted;" > "%RESULTS_DIR%\locks.txt"
echo Результат сохранен в файл: %RESULTS_DIR%\locks.txt

REM Проверка долгих запросов
echo Выполнение запроса: Проверка долгих запросов
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "SELECT pid, now() - pg_stat_activity.query_start AS duration, query, state FROM pg_stat_activity WHERE state != 'idle' AND now() - pg_stat_activity.query_start > interval '5 seconds' ORDER BY duration DESC;" > "%RESULTS_DIR%\long_queries.txt"
echo Результат сохранен в файл: %RESULTS_DIR%\long_queries.txt

REM Оптимизация таблицы refresh_tokens
echo Оптимизация таблицы refresh_tokens...
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "DELETE FROM refresh_tokens WHERE expires_at < NOW();"
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "DELETE FROM refresh_tokens WHERE id IN (SELECT id FROM (SELECT id, ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at DESC) as row_num FROM refresh_tokens) as ranked WHERE row_num > 5);"

REM Обновление статистики для оптимизатора запросов
echo Обновление статистики для оптимизатора запросов...
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "ANALYZE users;"
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "ANALYZE refresh_tokens;"
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "ANALYZE customer;"
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "ANALYZE courier;"
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "ANALYZE parcel;"
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -c "ANALYZE delivery;"

echo Анализ базы данных завершен. Результаты сохранены в директории: %RESULTS_DIR% 