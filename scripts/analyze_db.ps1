# Скрипт для анализа производительности базы данных (Windows)

$DB_HOST = "localhost"
$DB_PORT = "5432"
$DB_NAME = "delivery"
$DB_USER = "postgres"
$DB_PASSWORD = "postgres"

# Устанавливаем переменную окружения для пароля
$env:PGPASSWORD = $DB_PASSWORD

# Создаём папку для результатов
$RESULTS_DIR = "db_analysis_$(Get-Date -Format 'yyyyMMdd_HHmmss')"
New-Item -ItemType Directory -Path $RESULTS_DIR | Out-Null

# Функция для выполнения SQL-запроса и сохранения результата
function Run-Query {
    param ($query, $outputFile)
    
    Write-Host "Выполняется: $query"
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "$query" | Out-File -FilePath "$RESULTS_DIR\$outputFile" -Encoding utf8
}

# Анализ индексов
Run-Query "SELECT schemaname, tablename, indexname, indexdef FROM pg_indexes WHERE schemaname = 'public' ORDER BY tablename, indexname;" "indexes.txt"

# Анализ размера таблиц
Run-Query "SELECT relname AS table_name, pg_size_pretty(pg_total_relation_size(relid)) AS total_size FROM pg_catalog.pg_statio_user_tables ORDER BY pg_total_relation_size(relid) DESC;" "table_sizes.txt"

# Анализ производительности запросов
Run-Query "EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';" "query_users_email.txt"
Run-Query "EXPLAIN ANALYZE SELECT * FROM refresh_tokens WHERE user_id = 1;" "query_refresh_tokens_user_id.txt"

Write-Host "Анализ базы данных завершён! Результаты в $RESULTS_DIR"