package api

import (
	"delivery/internal/business/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CourierService interface {
	Create(courier *models.Courier) error
	Get(id int) (*models.Courier, error)
	Update(id int, courier *models.Courier) error
	Delete(id int) error
	List() ([]models.Courier, error)
	GetAvailableCouriers() ([]models.Courier, error)
	UpdateCourierStatus(id int, status string) error
}

type CourierHandler struct {
	service CourierService
}

func NewCourierHandler(service CourierService) *CourierHandler {
	return &CourierHandler{service: service}
}

func (h *CourierHandler) CreateCourier(w http.ResponseWriter, r *http.Request) {
	var courier models.Courier
	if err := json.NewDecoder(r.Body).Decode(&courier); err != nil {
		writeError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if err := h.service.Create(&courier); err != nil {
		writeError(w, "Failed to create courier", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(courier); err != nil {
		log.Printf("Ошибка при кодировании ответа: %v", err)
	}
}

func (h *CourierHandler) GetCourier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		writeError(w, "Invalid courier ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, "Invalid courier ID", http.StatusBadRequest)
		return
	}

	courier, err := h.service.Get(id)
	if err != nil {
		writeError(w, "Courier not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(courier); err != nil {
		log.Printf("Ошибка при кодировании ответа: %v", err)
	}
}

func (h *CourierHandler) UpdateCourier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, "Invalid courier ID", http.StatusBadRequest)
		return
	}

	var courier models.Courier
	if err := json.NewDecoder(r.Body).Decode(&courier); err != nil {
		writeError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if err := h.service.Update(id, &courier); err != nil {
		writeError(w, "Failed to update courier", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *CourierHandler) DeleteCourier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, "Invalid courier ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id); err != nil {
		writeError(w, "Failed to delete courier", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CourierHandler) ListCouriers(w http.ResponseWriter, r *http.Request) {
	couriers, err := h.service.List()
	if err != nil {
		writeError(w, "Failed to get couriers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(couriers); err != nil {
		log.Printf("Ошибка при кодировании ответа: %v", err)
	}
}

func (h *CourierHandler) GetAvailableCouriers(w http.ResponseWriter, r *http.Request) {
	couriers, err := h.service.GetAvailableCouriers()
	if err != nil {
		writeError(w, "Failed to fetch available couriers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(couriers)
}

func (h *CourierHandler) UpdateCourierStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, "Invalid courier ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateCourierStatus(id, input.Status); err != nil {
		writeError(w, "Failed to update courier status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
