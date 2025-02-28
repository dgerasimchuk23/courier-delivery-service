package api

import (
	"github.com/gorilla/mux"
)

// Маршрутизатор с зарегистрированными маршрутами
func NewRouter(parcelHandler *ParcelHandler, customerHandler *CustomerHandler, deliveryHandler *DeliveryHandler, courierHandler *CourierHandler) *mux.Router {

	r := mux.NewRouter()

	// Регистрирация маршрутов для посылок
	r.HandleFunc("/parcels", parcelHandler.CreateParcel).Methods("POST")
	r.HandleFunc("/parcels", parcelHandler.ListParcels).Methods("GET")
	r.HandleFunc("/parcels/{id}", parcelHandler.GetParcel).Methods("GET")
	r.HandleFunc("/parcels/{id}", parcelHandler.UpdateParcel).Methods("PUT")
	r.HandleFunc("/parcels/{id}/status", parcelHandler.UpdateParcelStatus).Methods("PUT")
	r.HandleFunc("/parcels/{id}/address", parcelHandler.UpdateParcelAddress).Methods("PUT")
	r.HandleFunc("/parcels/{id}", parcelHandler.DeleteParcel).Methods("DELETE")

	// Регистрирация маршрутов для клиентов
	r.HandleFunc("/customers", customerHandler.CreateCustomer).Methods("POST")
	r.HandleFunc("/customers", customerHandler.ListCustomers).Methods("GET")
	r.HandleFunc("/customers/{id}", customerHandler.GetCustomer).Methods("GET")
	r.HandleFunc("/customers/{id}", customerHandler.UpdateCustomer).Methods("PUT")
	r.HandleFunc("/customers/{id}", customerHandler.DeleteCustomer).Methods("DELETE")

	// Регистрирация маршрутов для доставок
	r.HandleFunc("/deliveries", deliveryHandler.CreateDelivery).Methods("POST")
	r.HandleFunc("/deliveries/assign", deliveryHandler.AssignDelivery).Methods("POST")
	r.HandleFunc("/deliveries/courier/{id}", deliveryHandler.GetDeliveriesByCourier).Methods("GET")
	r.HandleFunc("/deliveries/{id}", deliveryHandler.GetDelivery).Methods("GET")
	r.HandleFunc("/deliveries/{id}", deliveryHandler.UpdateDelivery).Methods("PUT")
	r.HandleFunc("/deliveries/{id}/complete", deliveryHandler.CompleteDelivery).Methods("PUT")
	r.HandleFunc("/deliveries/{id}", deliveryHandler.DeleteDelivery).Methods("DELETE")

	// Регистрирация маршрутов для курьеров
	r.HandleFunc("/couriers", courierHandler.CreateCourier).Methods("POST")
	r.HandleFunc("/couriers", courierHandler.ListCouriers).Methods("GET")
	r.HandleFunc("/couriers/available", courierHandler.GetAvailableCouriers).Methods("GET")
	r.HandleFunc("/couriers/{id}", courierHandler.GetCourier).Methods("GET")
	r.HandleFunc("/couriers/{id}", courierHandler.UpdateCourier).Methods("PUT")
	r.HandleFunc("/couriers/{id}/status", courierHandler.UpdateCourierStatus).Methods("PUT")
	r.HandleFunc("/couriers/{id}", courierHandler.DeleteCourier).Methods("DELETE")

	return r
}
