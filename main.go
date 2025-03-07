package main

import (
	"context"
	"delivery/config"
	"delivery/internal/api"
	"delivery/internal/auth"
	"delivery/internal/business/courier"
	"delivery/internal/business/customer"
	"delivery/internal/business/delivery"
	"delivery/internal/business/parcel"
	"delivery/internal/cache"
	"delivery/internal/db"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/lib/pq" // драйвер PostgreSQL
)

func main() {
	// Инициализация базы данных
	config, err := config.LoadConfig("./config/config.json")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	database := db.InitDB(config)
	defer db.CloseDB(database)

	// Инициализация Redis
	redisClient := cache.NewRedisClient(config)
	if redisClient != nil {
		defer redisClient.Close()
		log.Println("Redis успешно инициализирован")

		// Выводим начальную статистику Redis
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		stats, err := redisClient.MonitorStats(ctx)
		cancel()

		if err != nil {
			log.Printf("Ошибка при получении статистики Redis: %v", err)
		} else {
			log.Printf("Статистика Redis при запуске: total keys=%d, rate limit keys=%d, blacklist keys=%d",
				stats.TotalKeys, stats.RateLimitKeys, stats.BlacklistKeys)
			log.Printf("Память Redis: used=%d bytes, peak=%d bytes",
				stats.UsedMemory, stats.UsedMemoryPeak)
		}
	} else {
		log.Println("Не удалось инициализировать Redis, продолжаем без кэширования")
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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		stats, err := redisClient.MonitorStats(ctx)
		cancel()

		if err != nil {
			log.Printf("Ошибка при получении финальной статистики Redis: %v", err)
		} else {
			log.Printf("Финальная статистика Redis: total keys=%d, rate limit keys=%d, blacklist keys=%d",
				stats.TotalKeys, stats.RateLimitKeys, stats.BlacklistKeys)
			log.Printf("Память Redis: used=%d bytes, peak=%d bytes",
				stats.UsedMemory, stats.UsedMemoryPeak)
		}
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
