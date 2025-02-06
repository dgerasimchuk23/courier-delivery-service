# Используем базовый образ Go для сборки приложения
FROM golang:1.22.5 AS builder

# Устанавливаем рабочую директорию
WORKDIR /delivery

# Копируем файлы go.mod и go.sum и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Сборка приложения
RUN go build -o delivery main.go

# Используем минимальный образ для запуска собранного приложения
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /root/

# Создаём необходимые папки
RUN mkdir -p /root/config

# Копируем приложение и конфиги из предыдущего этапа сборки
COPY --from=builder /delivery/delivery .
COPY ./config ./config

# Запуск приложения
CMD ["./delivery"]
