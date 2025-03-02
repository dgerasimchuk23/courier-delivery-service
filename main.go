package main

import (
	"delivery/config"
	"delivery/internal/api"
	"delivery/internal/business/courier"
	"delivery/internal/business/customer"
	"delivery/internal/business/delivery"
	"delivery/internal/business/parcel"
	"delivery/internal/cache"
	"delivery/internal/db"
	"log"
	"strconv"

	_ "github.com/lib/pq" // драйвер PostgreSQL
)

func main() {
	// Инициализация базы данных
	config, err := config.LoadConfig("./config/config.json")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	database := db.InitDB(config)
	defer database.Close()

	// Инициализация Redis
	redisClient := cache.NewRedisClient(config)
	if redisClient != nil {
		defer redisClient.Close()
		log.Println("Redis успешно инициализирован")
	} else {
		log.Println("Не удалось инициализировать Redis, продолжаем без кэширования")
	}

	// Инициализация хранилищ
	customerStore := customer.NewCustomerStore(database)
	parcelStore := parcel.NewParcelStore(database)
	deliveryStore := delivery.NewDeliveryStore(database)
	courierStore := courier.NewCourierStore(database)

	// Инициализация сервисов
	customerService := customer.NewCustomerService(customerStore)
	parcelService := parcel.NewParcelService(parcelStore)
	deliveryService := delivery.NewDeliveryService(deliveryStore)
	courierService := courier.NewCourierService(courierStore)

	// Добавляем кэширование к сервисам, если Redis доступен
	if redisClient != nil {
		deliveryService.WithCache(redisClient)
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
	r := api.NewRouter(parcelHandler, customerHandler, deliveryHandler, courierHandler)

	// Запуск HTTP-сервера
	addr := config.Server.Host + ":" + strconv.Itoa(config.Server.Port)
	api.Start(addr, r)
}
