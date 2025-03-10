package main

import (
	"context"
	"delivery/internal/api"
	"delivery/internal/auth"
	"delivery/internal/business/courier"
	"delivery/internal/business/customer"
	"delivery/internal/business/delivery"
	"delivery/internal/business/parcel"
	"delivery/internal/cache"
	"delivery/internal/config"
	"delivery/internal/db"
	"delivery/internal/kafka"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Инициализация базы данных
	config, err := config.LoadConfig("./internal/config/config.json")
	if err != nil {
		log.Fatalf("Не удалось загрузить конфигурацию: %v", err)
	}

	database := db.InitDB(config)
	defer db.CloseDB(database)

	// Инициализация Redis
	redisClient := cache.NewRedisClient(config)
	if redisClient != nil {
		defer redisClient.Close()
		redisClient.LogInitialStats()
	} else {
		log.Println("Не удалось инициализировать Redis, продолжаем без кэширования")
	}

	// Инициализация Kafka
	kafkaClient, err := kafka.NewClient()
	if err != nil {
		log.Printf("Предупреждение: Не удалось инициализировать клиент Kafka: %v", err)
	} else {
		defer kafkaClient.Close()
		log.Println("Клиент Kafka успешно инициализирован")

		// Запускаем тестовый consumer и отправляем тестовое сообщение
		err = kafkaClient.StartTestConsumer("test-topic", "test-group")
		if err != nil {
			log.Printf("Предупреждение: Не удалось запустить тестовый consumer: %v", err)
		}

		// Отправляем тестовое сообщение через некоторое время
		go func() {
			time.Sleep(2 * time.Second)
			if err := kafkaClient.SendTestMessage("test-topic", "Привет, Kafka!"); err != nil {
				log.Printf("Предупреждение: Не удалось отправить тестовое сообщение: %v", err)
			}
		}()
	}

	// Инициализация хранилищ
	customerStore := customer.NewCustomerStore(database.DB)
	parcelStore := parcel.NewParcelStore(database.DB)
	deliveryStore := delivery.NewDeliveryStore(database.DB)
	courierStore := courier.NewCourierStore(database.DB)
	userStore := auth.NewUserStore(database.DB)

	// Инициализация сервисов
	customerService := customer.NewCustomerService(customerStore)
	parcelService := parcel.NewParcelService(parcelStore)
	deliveryService := delivery.NewDeliveryService(deliveryStore)
	courierService := courier.NewCourierService(courierStore)
	authService := auth.NewAuthService(userStore)

	// Закрываем ресурсы authService при завершении
	defer authService.Close()

	// Добавляем кэширование к сервисам, если Redis доступен
	if redisClient != nil {
		deliveryService.WithCache(redisClient)
		authService.WithCache(redisClient)
		// Для других сервисов можно добавить аналогично, когда они будут поддерживать кэширование
		// parcelService.WithCache(redisClient)
		// courierService.WithCache(redisClient)
	}

	// Инициализация обработчиков
	customerHandler := api.NewCustomerHandler(customerService)
	parcelHandler := api.NewParcelHandler(parcelService)
	deliveryHandler := api.NewDeliveryHandler(deliveryService)
	courierHandler := api.NewCourierHandler(courierService)

	// Создание маршрутизатора
	r := api.NewRouter(
		parcelHandler,
		customerHandler,
		deliveryHandler,
		courierHandler,
		authService,
		redisClient,
	)

	// Создание HTTP-сервера
	addr := config.Server.Host + ":" + strconv.Itoa(config.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Канал для получения сигналов завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск сервера в отдельной горутине
	go func() {
		log.Printf("Сервер запущен на %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	<-stop
	log.Println("Получен сигнал завершения, выполняется корректное завершение работы...")

	// Если Redis доступен, выводим финальную статистику
	if redisClient != nil {
		redisClient.LogFinalStats()
	}

	// Создаем контекст с таймаутом для корректного завершения
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Корректное завершение работы сервера
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при завершении работы сервера: %v", err)
	}

	log.Println("Сервер успешно остановлен")
}
