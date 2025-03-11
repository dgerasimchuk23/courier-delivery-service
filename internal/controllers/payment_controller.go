package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"delivery/internal/business/payment"
)

// PaymentController обрабатывает запросы, связанные с платежами
type PaymentController struct {
	paymentService payment.PaymentService
}

// NewPaymentController создает новый экземпляр PaymentController
func NewPaymentController() *PaymentController {
	return &PaymentController{
		paymentService: payment.NewMockPaymentService(),
	}
}

// CreatePayment обрабатывает запрос на создание нового платежа
func (pc *PaymentController) CreatePayment(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Декодируем тело запроса
	var req payment.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Создаем платеж
	response, err := pc.paymentService.CreatePayment(req)
	if err != nil {
		http.Error(w, "Ошибка создания платежа: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetPayment обрабатывает запрос на получение информации о платеже
func (pc *PaymentController) GetPayment(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID платежа из URL
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 {
		http.Error(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	paymentID := path[len(path)-1]

	// Получаем информацию о платеже
	response, err := pc.paymentService.GetPayment(paymentID)
	if err != nil {
		http.Error(w, "Ошибка получения информации о платеже: "+err.Error(), http.StatusNotFound)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CancelPayment обрабатывает запрос на отмену платежа
func (pc *PaymentController) CancelPayment(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID платежа из URL
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 4 || path[len(path)-1] != "cancel" {
		http.Error(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	paymentID := path[len(path)-2]

	// Отменяем платеж
	response, err := pc.paymentService.CancelPayment(paymentID)
	if err != nil {
		http.Error(w, "Ошибка отмены платежа: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RefundPayment обрабатывает запрос на возврат платежа
func (pc *PaymentController) RefundPayment(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID платежа из URL
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 4 || path[len(path)-1] != "refund" {
		http.Error(w, "Неверный URL", http.StatusBadRequest)
		return
	}
	paymentID := path[len(path)-2]

	// Декодируем тело запроса
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Возвращаем платеж
	response, err := pc.paymentService.RefundPayment(paymentID, req.Amount)
	if err != nil {
		http.Error(w, "Ошибка возврата платежа: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandlePayment обрабатывает все запросы, связанные с платежами
func (pc *PaymentController) HandlePayment(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Маршрутизация запросов
	switch {
	case strings.HasSuffix(path, "/payments") && r.Method == http.MethodPost:
		pc.CreatePayment(w, r)
	case strings.HasSuffix(path, "/cancel") && r.Method == http.MethodPost:
		pc.CancelPayment(w, r)
	case strings.HasSuffix(path, "/refund") && r.Method == http.MethodPost:
		pc.RefundPayment(w, r)
	case strings.Contains(path, "/payments/") && r.Method == http.MethodGet:
		pc.GetPayment(w, r)
	default:
		http.Error(w, "Неверный URL или метод", http.StatusNotFound)
	}
}

// RegisterRoutes регистрирует маршруты для платежного контроллера
func (pc *PaymentController) RegisterRoutes() {
	http.HandleFunc("/api/v1/payments", pc.CreatePayment)
	http.HandleFunc("/api/v1/payments/", pc.HandlePayment)
}
