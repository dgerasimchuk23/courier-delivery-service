package api

import (
	"delivery/internal/business/models"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ParcelService interface {
	Register(parcel *models.Parcel) error
	Get(id int) (*models.Parcel, error)
	Update(id int, parcel *models.Parcel) error
	UpdateStatus(id int, status string) error
	UpdateAddress(id int, address string) error
	Delete(id int) error
	List(clientID int) ([]models.Parcel, error)
}

type ParcelHandler struct {
	service ParcelService
}

func NewParcelHandler(service ParcelService) *ParcelHandler {
	return &ParcelHandler{service: service}
}

func (h *ParcelHandler) CreateParcel(w http.ResponseWriter, r *http.Request) {
	var parcel models.Parcel
	if err := json.NewDecoder(r.Body).Decode(&parcel); err != nil {
		writeError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if err := h.service.Register(&parcel); err != nil {
		writeError(w, "Failed to register parcel", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(parcel)
}

func (h *ParcelHandler) GetParcel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		writeError(w, "Missing parcel ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, "Invalid parcel ID", http.StatusBadRequest)
		return
	}

	parcel, err := h.service.Get(id)
	if err != nil {
		writeError(w, "Parcel not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parcel)
}

func (h *ParcelHandler) UpdateParcel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, "Invalid parcel ID", http.StatusBadRequest)
		return
	}

	var parcel models.Parcel
	if err := json.NewDecoder(r.Body).Decode(&parcel); err != nil {
		writeError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if err := h.service.Update(id, &parcel); err != nil {
		writeError(w, "Failed to update parcel", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ParcelHandler) UpdateParcelStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, "Invalid parcel ID", http.StatusBadRequest)
		return
	}

	var status struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		writeError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateStatus(id, status.Status); err != nil {
		writeError(w, "Failed to update status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ParcelHandler) UpdateParcelAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, "Invalid parcel ID", http.StatusBadRequest)
		return
	}

	var address struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&address); err != nil {
		writeError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateAddress(id, address.Address); err != nil {
		writeError(w, "Failed to update address", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ParcelHandler) DeleteParcel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeError(w, "Invalid parcel ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id); err != nil {
		writeError(w, "Failed to delete parcel", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ParcelHandler) ListParcels(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID, err := strconv.Atoi(vars["clientId"])
	if err != nil {
		writeError(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	parcels, err := h.service.List(clientID)
	if err != nil {
		writeError(w, "Failed to fetch parcels", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parcels)
}
