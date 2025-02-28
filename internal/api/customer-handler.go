package api

import (
	"delivery/internal/business/models"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CustomerService interface {
	Create(customer *models.Customer) error
	Get(id int) (*models.Customer, error)
	Update(id int, customer *models.Customer) error
	Delete(id int) error
	List() ([]models.Customer, error)
}

type CustomerHandler struct {
	service CustomerService
}

func NewCustomerHandler(service CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var customer models.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		writeError(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	if customer.Name == "" || customer.Email == "" {
		writeError(w, "Имя и email обязательны", http.StatusBadRequest)
		return
	}

	if err := h.service.Create(&customer); err != nil {
		writeError(w, "Не удалось создать клиента", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(customer)
}

func (h *CustomerHandler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		writeError(w, "Отсутствует ID клиента", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, "Некорректный ID клиента", http.StatusBadRequest)
		return
	}

	customer, err := h.service.Get(id)
	if err != nil {
		writeError(w, "Клиент не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

func (h *CustomerHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		writeError(w, "Отсутствует ID клиента", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, "Некорректный ID клиента", http.StatusBadRequest)
		return
	}

	var customer models.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		writeError(w, "Некорректные данные", http.StatusBadRequest)
		return
	}

	if err := h.service.Update(id, &customer); err != nil {
		writeError(w, "Не удалось обновить клиента", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *CustomerHandler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok || idStr == "" {
		writeError(w, "Отсутствует ID клиента", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, "Некорректный ID клиента", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id); err != nil {
		writeError(w, "Не удалось удалить клиента", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CustomerHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := h.service.List()
	if err != nil {
		writeError(w, "Не удалось получить список клиентов", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customers)
}
