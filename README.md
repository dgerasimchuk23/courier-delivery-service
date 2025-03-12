# Сервис курьерской доставки посылок

Данный проект представляет собой бэкенд-приложение для управления курьерской доставкой посылок. Проект разработан с применением современных практик разработки и включает в себя полный набор функций для управления доставками.

## Основные возможности

- **Управление клиентами**: Регистрация, обновление и удаление данных клиентов
- **Управление посылками**: Создание, отслеживание и обновление статуса посылок
- **Управление доставками**: Назначение курьеров, отслеживание статуса доставок
- **Платежная система**: Обработка платежей, возвраты и отмена платежей
- **Аутентификация**: JWT-based аутентификация для безопасного доступа к API
- **Мониторинг**: Интеграция с системами мониторинга (Prometheus, Grafana)

## Технический стек

- **Язык**: Go 1.24.0
- **База данных**: PostgreSQL
- **Кэширование**: Redis
- **Брокер сообщений**: Kafka
- **Протокол двусторонней связи**: WebSocket
- **Мониторинг**: Prometheus, Grafana
- **CI/CD**: GitHub Actions
- **Контейнеризация**: Docker

## Требования

- Go 1.24.0 или выше
- PostgreSQL
- Redis
- Kafka
- Docker и Docker Compose

## Установка и запуск

### Локальная разработка

1. Клонируйте репозиторий:
   ```sh
   git clone https://github.com/dgerasimchuk23/courier-delivery-service
   cd courier-delivery-service
   ```

2. Установите зависимости:
   ```sh
   go mod tidy
   ```

3. Создайте файл конфигурации:
   ```sh
   cp config/config.example.yaml config/config.yaml
   ```
   Отредактируйте config.yaml под свои нужды.

4. Запустите приложение:
   ```sh
   go run main.go
   ```

### Запуск через Docker

```sh
docker-compose up --build
```

## API Endpoints

### Клиенты
- `POST /api/v1/customers` - Создание клиента
- `GET /api/v1/customers` - Список клиентов
- `GET /api/v1/customers/{id}` - Получение клиента
- `PUT /api/v1/customers/{id}` - Обновление данных клиента
- `DELETE /api/v1/customers/{id}` - Удаление клиента

### Посылки
- `POST /api/v1/parcels` - Создание посылки
- `GET /api/v1/parcels` - Список посылок
- `GET /api/v1/parcels/{id}` - Получение посылки
- `PUT /api/v1/parcels/{id}` - Обновление посылки
- `PUT /api/v1/parcels/{id}/status` - Обновление статуса
- `DELETE /api/v1/parcels/{id}` - Удаление посылки

### Доставки
- `POST /api/v1/deliveries` - Создание доставки
- `GET /api/v1/deliveries` - Список доставок
- `GET /api/v1/deliveries/{id}` - Получение доставки
- `PUT /api/v1/deliveries/{id}` - Обновление доставки
- `PUT /api/v1/deliveries/{id}/status` - Обновление статуса
- `DELETE /api/v1/deliveries/{id}` - Удаление доставки

### Платежи
- `POST /api/v1/payments` - Создание нового платежа
- `GET /api/v1/payments/{id}` - Получение информации о платеже
- `POST /api/v1/payments/{id}/cancel` - Отмена платежа
- `POST /api/v1/payments/{id}/refund` - Возврат платежа

### Аутентификация
- `POST /api/v1/auth/register` - Регистрация пользователя
- `POST /api/v1/auth/login` - Вход в систему
- `POST /api/v1/auth/refresh` - Обновление токена
- `POST /api/v1/auth/logout` - Выход из системы

## Тестирование

### Локальный запуск тестов

```sh
go test -v ./...
```

### Запуск тестов через Docker

```sh
docker-compose run --rm test
```

Тесты запускаются в изолированном окружении с доступом к необходимым сервисам (PostgreSQL, Redis и Kafka), что обеспечивает стабильное и воспроизводимое выполнение тестов независимо от локального окружения.

## Мониторинг

Проект интегрирован с:
- Prometheus для сбора метрик
- Grafana для визуализации
- ELK Stack для логирования (в разработке)

## Платежная система

В проекте реализована платежная система с заглушкой (mock), которая эмулирует работу реальной платежной системы. Это позволяет тестировать интеграцию платежного функционала без необходимости подключения к реальным платежным сервисам.

### Возможности платежной системы

- **Методы оплаты**:
  - Банковские карты
  - Банковские переводы
  - Электронные кошельки

- **Операции**:
  - Создание платежей
  - Получение информации о платеже
  - Отмена платежей
  - Возврат средств
  - Отслеживание статуса платежа

- **Статусы платежей**:
  - `pending` - в ожидании
  - `completed` - завершен
  - `failed` - отменен
  - `refunded` - возвращен

### Пример использования

```go
// Создание сервиса
paymentService := payment.NewMockPaymentService()

// Создание платежа
request := payment.PaymentRequest{
    OrderID:     "order_123",
    Amount:      100.00,
    Currency:    "RUB",
    Method:      payment.MethodCard,
    Description: "Оплата заказа #123",
    CustomerInfo: payment.CustomerInfo{
        Name:  "Иван Иванов",
        Email: "ivan@example.com",
        Phone: "+7 (999) 123-45-67",
    },
}

response, err := paymentService.CreatePayment(request)
if err != nil {
    log.Fatalf("Ошибка создания платежа: %v", err)
}

// Получение информации о платеже
payment, err := paymentService.GetPayment(response.PaymentID)
if err != nil {
    log.Fatalf("Ошибка получения информации о платеже: %v", err)
}

// Отмена платежа
canceledPayment, err := paymentService.CancelPayment(response.PaymentID)
if err != nil {
    log.Fatalf("Ошибка отмены платежа: %v", err)
}
```

## Лицензия

MIT

## Автор

Дмитрий Герасимчук
