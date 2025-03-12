package api

import (
	"context"
	"delivery/internal/auth"
	"delivery/internal/cache"
	"delivery/internal/middleware"

	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Маршрутизатор с зарегистрированными маршрутами
func NewRouter(
	parcelHandler *ParcelHandler,
	customerHandler *CustomerHandler,
	deliveryHandler *DeliveryHandler,
	courierHandler *CourierHandler,
	authService *auth.AuthService,
	redisClient *cache.RedisClient,
	wsManager *WebSocketManager,
) *mux.Router {

	r := mux.NewRouter()

	// Инициализация middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)
	rateLimiter := middleware.NewRateLimiter(redisClient, middleware.DefaultRateLimitConfig())

	// Загружаем конфигурацию Rate Limiting из Redis
	ctx := context.Background()
	if err := rateLimiter.LoadConfigFromRedis(ctx); err != nil {
		// Если не удалось загрузить конфигурацию, используем значения по умолчанию
		// и сохраняем их в Redis
		if err := rateLimiter.SaveConfigToRedis(ctx); err != nil {
			log.Printf("Ошибка при сохранении конфигурации Rate Limiting в Redis: %v", err)
		}
	}

	// Применяем middleware ко всем маршрутам
	r.Use(middleware.MetricsMiddleware)
	r.Use(authMiddleware.Middleware())
	r.Use(rateLimiter.Middleware())

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

	// Добавляем маршрут для WebSocket соединений
	// Этот маршрут не требует аутентификации, поэтому добавляем его отдельно
	wsRouter := r.PathPrefix("/ws").Subrouter()
	wsRouter.Use(rateLimiter.Middleware()) // Применяем только ограничение скорости
	wsRouter.HandleFunc("/orders", wsManager.WebSocketHandler)

	// Добавляем маршрут для обновления конфигурации Rate Limiting
	r.HandleFunc("/admin/rate-limit", func(w http.ResponseWriter, r *http.Request) {
		// Этот маршрут должен быть защищен дополнительной аутентификацией
		// TODO: Добавить проверку прав администратора

		var config middleware.RateLimitConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, "Неверный формат данных", http.StatusBadRequest)
			return
		}

		if err := rateLimiter.UpdateConfig(config); err != nil {
			http.Error(w, "Ошибка при обновлении конфигурации", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
			log.Printf("Ошибка при кодировании ответа: %v", err)
		}
	}).Methods("POST")

	// Добавляем эндпоинт для метрик Prometheus
	// Этот эндпоинт не требует аутентификации
	metricsRouter := r.PathPrefix("/metrics").Subrouter()
	metricsRouter.Handle("", promhttp.Handler())

	return r
}
