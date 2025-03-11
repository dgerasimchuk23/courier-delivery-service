# Используем базовый образ Go для сборки приложения
FROM golang:1.24.0 AS builder

# Устанавливаем рабочую директорию
WORKDIR /delivery

# Копируем файлы go.mod и go.sum и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 go build -o delivery main.go

# Используем минимальный образ для запуска собранного приложения
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk add --no-cache libc6-compat

# Устанавливаем рабочую директорию
WORKDIR /root/

# Создаём необходимые папки
RUN mkdir -p /root/config

# Устанавливаем переменную окружения для определения, что приложение запущено в контейнере
ENV IN_CONTAINER=true

# Копируем приложение и конфиги из предыдущего этапа сборки
COPY --from=builder /delivery/delivery .
COPY ./config ./config

# Запуск приложения
CMD ["./delivery"]
