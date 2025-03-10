# Сервис курьерской доставки посылок

Этот проект представляет собой бэкенд-приложение для управления курьерской доставкой посылок.

## Возможности

- **Клиенты**: Управление данными клиентов.
- **Посылки**: Отслеживание информации и статуса посылок.
- **Доставки**: Назначение и обновление статусов доставок.

## Требования

- Go 1.22.5
- SQLite
- Docker

## Установка

1. Клонируйте репозиторий:
   ```sh
   git clone https://github.com/dgerasimchuk23/courier-delivery-service
   cd delivery
   ```

2. Настройте базу данных:
   - Убедитесь, что `db/schema.go` соответствует вашей схеме базы данных.
   - Проверьте `db/db.go` и убедитесь, что в нём правильные параметры подключения.

3. Установите зависимости:
   ```sh
   go mod tidy
   ```

4. Запустите приложение:
   ```sh
   go run main.go
   ```

## Документация API

API предоставляет следующие **методы**:

- **Клиенты**
  - `POST /customers` — создание клиента
  - `GET /customers` — список клиентов
  - `GET /customers/{id}` — получить клиента
  - `PUT /customers/{id}` — обновить данные клиента
  - `DELETE /customers/{id}` — удалить клиента

- **Посылки**
  - `POST /parcels` — создание посылки
  - `GET /parcels` — список посылок
  - `GET /parcels/{id}` — получить посылку
  - `PUT /parcels/{id}` — обновить посылку
  - `PUT /parcels/{id}/status` — обновить статус посылки
  - `PUT /parcels/{id}/address` — обновить адрес посылки
  - `DELETE /parcels/{id}` — удалить посылку

- **Доставки**
  - `POST /deliveries` — создание доставки
  - `POST /deliveries/assign` — назначение доставки курьеру
  - `GET /deliveries/courier/{id}` — получить доставки курьера
  - `GET /deliveries/{id}` — получить доставку
  - `PUT /deliveries/{id}` — обновить доставку
  - `PUT /deliveries/{id}/complete` — завершить доставку
  - `DELETE /deliveries/{id}` — удалить доставку

Подробнее смотрите в папке `api/`.

## Развёртывание с использованием Docker

Запустите команду:
```sh
docker-compose up --build
```
После этого API будет доступен по адресу: [http://localhost:8080](http://localhost:8080)

Готовый образ доступен на Docker Hub:
[docker.io/sprut13/delivery](https://hub.docker.com/repository/docker/sprut13/delivery/general)

## Тестирование

- Модульные тесты находятся рядом с соответствующими модулями.
- Запуск тестов:
  ```sh
  go test -v ./...
  ```
