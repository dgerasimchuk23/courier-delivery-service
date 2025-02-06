package main

import (
	"delivery/config"
	"delivery/internal/api"
	"delivery/internal/customer"
	"delivery/internal/db"
	"delivery/internal/delivery"
	"delivery/internal/parcel"
	"log"
	"strconv"

	_ "modernc.org/sqlite"
)

func main() {
	// Инициализация базы данных
	config, err := config.LoadConfig("./config/config.json") // Убедитесь, что путь корректный
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	database := db.InitDB(config.Database.DSN)
	defer database.Close()

	// Инициализация хранилищ
	customerStore := customer.NewCustomerStore(database)
	parcelStore := parcel.NewParcelStore(database)
	deliveryStore := delivery.NewDeliveryStore(database)

	// Инициализация сервисов
	customerService := customer.NewCustomerService(customerStore)
	parcelService := parcel.NewParcelService(parcelStore)
	deliveryService := delivery.NewDeliveryService(deliveryStore)

	// Инициализация обработчиков
	customerHandler := api.NewCustomerHandler(customerService)
	parcelHandler := api.NewParcelHandler(parcelService)
	deliveryHandler := api.NewDeliveryHandler(deliveryService)

	// Создание маршрутизатора
	r := api.NewRouter(parcelHandler, customerHandler, deliveryHandler)

	// Запуск HTTP-сервера
	addr := config.Server.Host + ":" + strconv.Itoa(config.Server.Port)
	api.Start(addr, r)
}
