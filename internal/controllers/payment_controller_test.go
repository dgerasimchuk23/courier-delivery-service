package controllers

import (
	"bytes"
	"delivery/internal/business/payment"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockPaymentService реализует интерфейс payment.PaymentService для тестирования
type MockPaymentService struct {
	CreatePaymentFunc func(request payment.PaymentRequest) (*payment.PaymentResponse, error)
	GetPaymentFunc    func(paymentID string) (*payment.PaymentResponse, error)
	CancelPaymentFunc func(paymentID string) (*payment.PaymentResponse, error)
	RefundPaymentFunc func(paymentID string, amount float64) (*payment.PaymentResponse, error)
	ChargeFunc        func(amount float64, currency string) (string, error)
}

func (m *MockPaymentService) CreatePayment(request payment.PaymentRequest) (*payment.PaymentResponse, error) {
	return m.CreatePaymentFunc(request)
}

func (m *MockPaymentService) GetPayment(paymentID string) (*payment.PaymentResponse, error) {
	return m.GetPaymentFunc(paymentID)
}

func (m *MockPaymentService) CancelPayment(paymentID string) (*payment.PaymentResponse, error) {
	return m.CancelPaymentFunc(paymentID)
}

func (m *MockPaymentService) RefundPayment(paymentID string, amount float64) (*payment.PaymentResponse, error) {
	return m.RefundPaymentFunc(paymentID, amount)
}

func (m *MockPaymentService) Charge(amount float64, currency string) (string, error) {
	return m.ChargeFunc(amount, currency)
}

func TestCreatePaymentHandler(t *testing.T) {
	// Создаем мок сервиса
	mockService := &MockPaymentService{
		CreatePaymentFunc: func(request payment.PaymentRequest) (*payment.PaymentResponse, error) {
			return &payment.PaymentResponse{
				PaymentID: "test_payment_id",
				OrderID:   request.OrderID,
				Amount:    request.Amount,
				Currency:  request.Currency,
				Status:    payment.StatusPending,
				Method:    request.Method,
			}, nil
		},
	}

	// Создаем контроллер с мок сервисом
	controller := &PaymentController{
		paymentService: mockService,
	}

	// Создаем тестовый запрос
	requestBody := `{
		"order_id": "order_123",
		"amount": 100.00,
		"currency": "RUB",
		"method": "card",
		"description": "Test payment",
		"customer_info": {
			"name": "Test User",
			"email": "test@example.com",
			"phone": "+7 (999) 123-45-67"
		}
	}`

	req, err := http.NewRequest("POST", "/api/v1/payments", bytes.NewBufferString(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	controller.CreatePayment(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Проверяем тело ответа
	var response payment.PaymentResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	if response.PaymentID != "test_payment_id" {
		t.Errorf("handler returned unexpected payment ID: got %v want %v", response.PaymentID, "test_payment_id")
	}
}

func TestGetPaymentHandler(t *testing.T) {
	// Создаем мок сервиса
	mockService := &MockPaymentService{
		GetPaymentFunc: func(paymentID string) (*payment.PaymentResponse, error) {
			if paymentID == "test_payment_id" {
				return &payment.PaymentResponse{
					PaymentID: paymentID,
					OrderID:   "order_123",
					Amount:    100.00,
					Currency:  "RUB",
					Status:    payment.StatusPending,
					Method:    payment.MethodCard,
				}, nil
			}
			return nil, payment.ErrPaymentNotFound
		},
	}

	// Создаем контроллер с мок сервисом
	controller := &PaymentController{
		paymentService: mockService,
	}

	// Создаем тестовый запрос
	req, err := http.NewRequest("GET", "/api/v1/payments/test_payment_id", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Имитируем извлечение ID из URL
		r.URL.Path = "/api/v1/payments/test_payment_id"
		controller.GetPayment(w, r)
	})

	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Проверяем тело ответа
	var response payment.PaymentResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	if response.PaymentID != "test_payment_id" {
		t.Errorf("handler returned unexpected payment ID: got %v want %v", response.PaymentID, "test_payment_id")
	}
}

func TestCancelPaymentHandler(t *testing.T) {
	// Создаем мок сервиса
	mockService := &MockPaymentService{
		CancelPaymentFunc: func(paymentID string) (*payment.PaymentResponse, error) {
			if paymentID == "test_payment_id" {
				return &payment.PaymentResponse{
					PaymentID:    paymentID,
					OrderID:      "order_123",
					Amount:       100.00,
					Currency:     "RUB",
					Status:       payment.StatusFailed,
					Method:       payment.MethodCard,
					ErrorMessage: "Платеж отменен пользователем",
				}, nil
			}
			return nil, payment.ErrPaymentNotFound
		},
	}

	// Создаем контроллер с мок сервисом
	controller := &PaymentController{
		paymentService: mockService,
	}

	// Создаем тестовый запрос
	req, err := http.NewRequest("POST", "/api/v1/payments/test_payment_id/cancel", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Имитируем извлечение ID из URL
		r.URL.Path = "/api/v1/payments/test_payment_id/cancel"
		controller.CancelPayment(w, r)
	})

	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Проверяем тело ответа
	var response payment.PaymentResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	if response.Status != payment.StatusFailed {
		t.Errorf("handler returned unexpected status: got %v want %v", response.Status, payment.StatusFailed)
	}
}

func TestRefundPaymentHandler(t *testing.T) {
	// Создаем мок сервиса
	mockService := &MockPaymentService{
		RefundPaymentFunc: func(paymentID string, amount float64) (*payment.PaymentResponse, error) {
			if paymentID == "test_payment_id" {
				return &payment.PaymentResponse{
					PaymentID: paymentID,
					OrderID:   "order_123",
					Amount:    100.00,
					Currency:  "RUB",
					Status:    payment.StatusRefunded,
					Method:    payment.MethodCard,
				}, nil
			}
			return nil, payment.ErrPaymentNotFound
		},
	}

	// Создаем контроллер с мок сервисом
	controller := &PaymentController{
		paymentService: mockService,
	}

	// Создаем тестовый запрос
	requestBody := `{"amount": 50.00}`
	req, err := http.NewRequest("POST", "/api/v1/payments/test_payment_id/refund", bytes.NewBufferString(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Имитируем извлечение ID из URL
		r.URL.Path = "/api/v1/payments/test_payment_id/refund"
		controller.RefundPayment(w, r)
	})

	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Проверяем тело ответа
	var response payment.PaymentResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	if response.Status != payment.StatusRefunded {
		t.Errorf("handler returned unexpected status: got %v want %v", response.Status, payment.StatusRefunded)
	}
}
