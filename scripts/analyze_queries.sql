-- Анализ и оптимизация запросов для базы данных delivery

-- 1. Анализ индексов
SELECT
    schemaname,
    tablename,
    indexname,
    indexdef
FROM
    pg_indexes
WHERE
    schemaname = 'public'
ORDER BY
    tablename,
    indexname;

-- 2. Анализ размера таблиц
SELECT
    relname AS table_name,
    pg_size_pretty(pg_total_relation_size(relid)) AS total_size,
    pg_size_pretty(pg_relation_size(relid)) AS table_size,
    pg_size_pretty(pg_total_relation_size(relid) - pg_relation_size(relid)) AS index_size
FROM
    pg_catalog.pg_statio_user_tables
ORDER BY
    pg_total_relation_size(relid) DESC;

-- 3. Анализ запросов для таблицы users
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';
EXPLAIN ANALYZE SELECT * FROM users WHERE id = 1;

-- 4. Анализ запросов для таблицы refresh_tokens
EXPLAIN ANALYZE SELECT * FROM refresh_tokens WHERE token = 'test-token';
EXPLAIN ANALYZE SELECT * FROM refresh_tokens WHERE user_id = 1;
EXPLAIN ANALYZE SELECT * FROM refresh_tokens WHERE expires_at < NOW();

-- 5. Анализ запросов для таблицы customer
EXPLAIN ANALYZE SELECT * FROM customer WHERE email = 'customer@example.com';
EXPLAIN ANALYZE SELECT * FROM customer WHERE id = 1;

-- 6. Анализ запросов для таблицы courier
EXPLAIN ANALYZE SELECT * FROM courier WHERE email = 'courier@example.com';
EXPLAIN ANALYZE SELECT * FROM courier WHERE status = 'active';

-- 7. Анализ запросов для таблицы parcel
EXPLAIN ANALYZE SELECT * FROM parcel WHERE client = 1;
EXPLAIN ANALYZE SELECT * FROM parcel WHERE status = 'pending';

-- 8. Анализ запросов для таблицы delivery
EXPLAIN ANALYZE SELECT * FROM delivery WHERE courier_id = 1;
EXPLAIN ANALYZE SELECT * FROM delivery WHERE parcel_id = 1;
EXPLAIN ANALYZE SELECT * FROM delivery WHERE status = 'delivered';

-- 9. Анализ сложных запросов с JOIN
EXPLAIN ANALYZE
SELECT d.*, p.address, c.name AS courier_name, cu.name AS customer_name
FROM delivery d
JOIN parcel p ON d.parcel_id = p.id
JOIN courier c ON d.courier_id = c.id
JOIN customer cu ON p.client = cu.id
WHERE d.status = 'delivered';

-- 10. Оптимизация таблицы refresh_tokens
-- Удаление просроченных токенов
DELETE FROM refresh_tokens WHERE expires_at < NOW();

-- Удаление дубликатов токенов, оставляя только последние 5 для каждого пользователя
DELETE FROM refresh_tokens 
WHERE id IN (
    SELECT id FROM (
        SELECT id, 
               ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at DESC) as row_num 
        FROM refresh_tokens
    ) as ranked 
    WHERE row_num > 5
);

-- 11. Обновление статистики для оптимизатора запросов
ANALYZE users;
ANALYZE refresh_tokens;
ANALYZE customer;
ANALYZE courier;
ANALYZE parcel;
ANALYZE delivery;

-- 12. Проверка блокировок
SELECT 
    blocked_locks.pid AS blocked_pid,
    blocking_locks.pid AS blocking_pid,
    blocked_activity.usename AS blocked_user,
    blocking_activity.usename AS blocking_user,
    blocked_activity.query AS blocked_statement,
    blocking_activity.query AS blocking_statement
FROM 
    pg_catalog.pg_locks blocked_locks
JOIN 
    pg_catalog.pg_locks blocking_locks 
    ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid
JOIN 
    pg_catalog.pg_stat_activity blocked_activity
    ON blocked_activity.pid = blocked_locks.pid
JOIN 
    pg_catalog.pg_stat_activity blocking_activity
    ON blocking_activity.pid = blocking_locks.pid
WHERE 
    NOT blocked_locks.granted;

-- 13. Проверка долгих запросов
SELECT 
    pid,
    now() - pg_stat_activity.query_start AS duration,
    query,
    state
FROM 
    pg_stat_activity
WHERE 
    state != 'idle'
    AND now() - pg_stat_activity.query_start > interval '5 seconds'
ORDER BY 
    duration DESC; 