package api

import (
	"delivery/internal/models"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type DeliveryService interface {
	Create(delivery *models.Delivery) error
	Get(id int) (*models.Delivery, error)
	Update(id int, delivery *models.Delivery) error
	Delete(id int) error
	GetByParcelID(parcelID int) (*models.Delivery, error)
	AssignDelivery(courierID, parcelID int) (models.Delivery, error)
	CompleteDelivery(deliveryID int) error
	GetDeliveriesByCourier(courierID int) ([]models.Delivery, error)
}

type DeliveryHandler struct {
	service DeliveryService
}

func NewDeliveryHandler(service DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{service: service}
}

func (h *DeliveryHandler) CreateDelivery(w http.ResponseWriter, r *http.Request) {
	var delivery models.Delivery
	if err := json.NewDecoder(r.Body).Decode(&delivery); err != nil {
		writeError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if err := h.service.Create(&delivery); err != nil {
		writeError(w, "Failed to create delivery", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(delivery)
}

func (h *DeliveryHandler) GetDelivery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		writeError(w, "Missing delivery ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, "Invalid delivery ID", http.StatusBadRequest)
		return
	}

	delivery, err := h.service.Get(id)
	if err != nil {
		writeError(w, "Delivery not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(delivery)
}

func (h *DeliveryHandler) UpdateDelivery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, "Invalid delivery ID", http.StatusBadRequest)
		return
	}

	var delivery models.Delivery
	if err := json.NewDecoder(r.Body).Decode(&delivery); err != nil {
		writeError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if err := h.service.Update(id, &delivery); err != nil {
		writeError(w, "Failed to update delivery", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *DeliveryHandler) DeleteDelivery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, "Invalid delivery ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id); err != nil {
		writeError(w, "Failed to delete delivery", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DeliveryHandler) AssignDelivery(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CourierID int `json:"courier_id"`
		ParcelID  int `json:"parcel_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	delivery, err := h.service.AssignDelivery(input.CourierID, input.ParcelID)
	if err != nil {
		writeError(w, "Failed to assign delivery", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(delivery)
}

func (h *DeliveryHandler) CompleteDelivery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deliveryID, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, "Invalid delivery ID", http.StatusBadRequest)
		return
	}

	if err := h.service.CompleteDelivery(deliveryID); err != nil {
		writeError(w, "Failed to complete delivery", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *DeliveryHandler) GetDeliveriesByCourier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courierID, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, "Invalid courier ID", http.StatusBadRequest)
		return
	}

	deliveries, err := h.service.GetDeliveriesByCourier(courierID)
	if err != nil {
		writeError(w, "Failed to fetch deliveries", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deliveries)
}
