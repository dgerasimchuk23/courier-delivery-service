package main

import (
	"delivery/config"
	"delivery/internal/api"
	"delivery/internal/business/courier"
	"delivery/internal/business/customer"
	"delivery/internal/business/delivery"
	"delivery/internal/business/parcel"
	"delivery/internal/db"
	"log"
	"strconv"

	_ "modernc.org/sqlite"
)

func main() {
	// Инициализация базы данных
	config, err := config.LoadConfig("./config/config.json")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	database := db.InitDB(config.Database.DSN)
	defer database.Close()

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
