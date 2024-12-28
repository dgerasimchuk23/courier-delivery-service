package main

import (
	"delivery/internal/customer"
	"delivery/internal/parcel"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	// Инициализация баз данных
	parcelsDB, err := parcel.SetupParcelsDB()
	if err != nil {
		log.Printf("Ошибка при настройке базы данных parcels: %v", err)
		return
	}
	defer parcelsDB.Close()

	customersDB := customer.SetupCustomersDB()
	defer customersDB.Close()

	// Инициализация сервисов
	parcelStore := parcel.NewParcelStore(parcelsDB)
	parcelService := parcel.NewParcelService(parcelStore)
	fmt.Printf("Сервис %v инициализирован и готов к работе\n", parcelService)

	customerStore := customer.NewCustomerStore(customersDB)
	customerService := customer.NewCustomerService(customerStore)
	fmt.Printf("Сервис %v инициализирован и готов к работе\n", customerService)

}
