package payment

import (
	"errors"
	"fmt"
	"time"
)

// Ошибки платежной системы
var (
	ErrPaymentNotFound = errors.New("платеж не найден")
)

// PaymentStatus представляет статус платежа
type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusCompleted PaymentStatus = "completed"
	StatusFailed    PaymentStatus = "failed"
	StatusRefunded  PaymentStatus = "refunded"
)

// PaymentMethod представляет метод оплаты
type PaymentMethod string

const (
	MethodCard             PaymentMethod = "card"
	MethodBankTransfer     PaymentMethod = "bank_transfer"
	MethodElectronicWallet PaymentMethod = "electronic_wallet"
)

// PaymentRequest представляет запрос на оплату
type PaymentRequest struct {
	OrderID      string        `json:"order_id"`
	Amount       float64       `json:"amount"`
	Currency     string        `json:"currency"`
	Method       PaymentMethod `json:"method"`
	Description  string        `json:"description"`
	CustomerInfo CustomerInfo  `json:"customer_info"`
}

// CustomerInfo представляет информацию о клиенте
type CustomerInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// PaymentResponse представляет ответ на запрос оплаты
type PaymentResponse struct {
	PaymentID    string        `json:"payment_id"`
	OrderID      string        `json:"order_id"`
	Amount       float64       `json:"amount"`
	Currency     string        `json:"currency"`
	Status       PaymentStatus `json:"status"`
	Method       PaymentMethod `json:"method"`
	CreatedAt    time.Time     `json:"created_at"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
	RedirectURL  string        `json:"redirect_url,omitempty"`
	ReceiptURL   string        `json:"receipt_url,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// PaymentService - интерфейс для работы с платежами
// Charge - метод для обработки платежа
// amount - сумма платежа
// currency - валюта платежа
// Возвращает строку с результатом и ошибку, если она возникла

type PaymentService interface {
	// Создает новый платеж
	CreatePayment(request PaymentRequest) (*PaymentResponse, error)

	// Получает информацию о платеже по ID
	GetPayment(paymentID string) (*PaymentResponse, error)

	// Отменяет платеж
	CancelPayment(paymentID string) (*PaymentResponse, error)

	// Возвращает платеж
	RefundPayment(paymentID string, amount float64) (*PaymentResponse, error)

	// Обрабатывает платеж (устаревший метод, сохранен для обратной совместимости)
	Charge(amount float64, currency string) (string, error)
}

// MockPaymentService - заглушка для платежного сервиса
// Charge - имитация процесса оплаты
// amount - сумма платежа
// currency - валюта платежа
// Возвращает строку с результатом и ошибку, если она возникла

type MockPaymentService struct {
	// Хранилище платежей для имитации базы данных
	payments map[string]*PaymentResponse
}

// NewMockPaymentService создает новый экземпляр MockPaymentService
func NewMockPaymentService() *MockPaymentService {
	return &MockPaymentService{
		payments: make(map[string]*PaymentResponse),
	}
}

// CreatePayment создает новый платеж
func (m *MockPaymentService) CreatePayment(request PaymentRequest) (*PaymentResponse, error) {
	// Валидация запроса
	if request.Amount <= 0 {
		return nil, fmt.Errorf("недопустимая сумма платежа")
	}

	if request.Currency == "" {
		return nil, fmt.Errorf("валюта не указана")
	}

	// Генерация ID платежа (в реальном сервисе был бы более сложный алгоритм)
	paymentID := fmt.Sprintf("pay_%d", time.Now().UnixNano())

	// Создание ответа
	now := time.Now()
	response := &PaymentResponse{
		PaymentID:   paymentID,
		OrderID:     request.OrderID,
		Amount:      request.Amount,
		Currency:    request.Currency,
		Status:      StatusPending,
		Method:      request.Method,
		CreatedAt:   now,
		RedirectURL: fmt.Sprintf("https://example.com/payments/%s", paymentID),
	}

	// Сохранение платежа в "базе данных"
	m.payments[paymentID] = response

	return response, nil
}

// GetPayment получает информацию о платеже по ID
func (m *MockPaymentService) GetPayment(paymentID string) (*PaymentResponse, error) {
	payment, exists := m.payments[paymentID]
	if !exists {
		return nil, ErrPaymentNotFound
	}

	return payment, nil
}

// CancelPayment отменяет платеж
func (m *MockPaymentService) CancelPayment(paymentID string) (*PaymentResponse, error) {
	payment, exists := m.payments[paymentID]
	if !exists {
		return nil, ErrPaymentNotFound
	}

	if payment.Status != StatusPending {
		return nil, fmt.Errorf("нельзя отменить платеж в статусе %s", payment.Status)
	}

	payment.Status = StatusFailed
	payment.ErrorMessage = "Платеж отменен пользователем"

	return payment, nil
}

// RefundPayment возвращает платеж
func (m *MockPaymentService) RefundPayment(paymentID string, amount float64) (*PaymentResponse, error) {
	payment, exists := m.payments[paymentID]
	if !exists {
		return nil, ErrPaymentNotFound
	}

	if payment.Status != StatusCompleted {
		return nil, fmt.Errorf("нельзя вернуть платеж в статусе %s", payment.Status)
	}

	if amount > payment.Amount {
		return nil, fmt.Errorf("сумма возврата не может превышать сумму платежа")
	}

	payment.Status = StatusRefunded
	now := time.Now()
	payment.CompletedAt = &now

	return payment, nil
}

// Charge обрабатывает платеж (устаревший метод, сохранен для обратной совместимости)
func (m *MockPaymentService) Charge(amount float64, currency string) (string, error) {
	if amount <= 0 {
		return "", fmt.Errorf("недопустимая сумма платежа")
	}

	// Создаем платеж через новый интерфейс
	request := PaymentRequest{
		OrderID:     fmt.Sprintf("order_%d", time.Now().Unix()),
		Amount:      amount,
		Currency:    currency,
		Method:      MethodCard,
		Description: "Оплата через устаревший метод Charge",
	}

	response, err := m.CreatePayment(request)
	if err != nil {
		return "", err
	}

	// Имитируем успешное завершение платежа
	now := time.Now()
	response.Status = StatusCompleted
	response.CompletedAt = &now
	response.ReceiptURL = fmt.Sprintf("https://example.com/receipts/%s", response.PaymentID)

	return fmt.Sprintf("Платеж на сумму %.2f %s успешно обработан. ID платежа: %s", amount, currency, response.PaymentID), nil
}
