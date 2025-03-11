package payment

import (
	"testing"
	"time"
)

func TestNewMockPaymentService(t *testing.T) {
	service := NewMockPaymentService()
	if service == nil {
		t.Fatal("Expected non-nil service")
	}
	if service.payments == nil {
		t.Fatal("Expected non-nil payments map")
	}
}

func TestCreatePayment(t *testing.T) {
	service := NewMockPaymentService()

	// Тест успешного создания платежа
	request := PaymentRequest{
		OrderID:     "order_123",
		Amount:      100.00,
		Currency:    "RUB",
		Method:      MethodCard,
		Description: "Test payment",
		CustomerInfo: CustomerInfo{
			Name:  "Test User",
			Email: "test@example.com",
			Phone: "+7 (999) 123-45-67",
		},
	}

	response, err := service.CreatePayment(request)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.PaymentID == "" {
		t.Fatal("Expected non-empty payment ID")
	}

	if response.OrderID != request.OrderID {
		t.Errorf("Expected OrderID %s, got %s", request.OrderID, response.OrderID)
	}

	if response.Amount != request.Amount {
		t.Errorf("Expected Amount %.2f, got %.2f", request.Amount, response.Amount)
	}

	if response.Currency != request.Currency {
		t.Errorf("Expected Currency %s, got %s", request.Currency, response.Currency)
	}

	if response.Status != StatusPending {
		t.Errorf("Expected Status %s, got %s", StatusPending, response.Status)
	}

	// Тест с некорректной суммой
	invalidRequest := PaymentRequest{
		OrderID:  "order_456",
		Amount:   -100.00, // Отрицательная сумма
		Currency: "RUB",
		Method:   MethodCard,
	}

	_, err = service.CreatePayment(invalidRequest)
	if err == nil {
		t.Fatal("Expected error for negative amount, got nil")
	}

	// Тест с пустой валютой
	emptyRequest := PaymentRequest{
		OrderID:  "order_789",
		Amount:   100.00,
		Currency: "", // Пустая валюта
		Method:   MethodCard,
	}

	_, err = service.CreatePayment(emptyRequest)
	if err == nil {
		t.Fatal("Expected error for empty currency, got nil")
	}
}

func TestGetPayment(t *testing.T) {
	service := NewMockPaymentService()

	// Создаем платеж для тестирования
	request := PaymentRequest{
		OrderID:  "order_123",
		Amount:   100.00,
		Currency: "RUB",
		Method:   MethodCard,
	}

	response, _ := service.CreatePayment(request)
	paymentID := response.PaymentID

	// Тест успешного получения платежа
	payment, err := service.GetPayment(paymentID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if payment.PaymentID != paymentID {
		t.Errorf("Expected PaymentID %s, got %s", paymentID, payment.PaymentID)
	}

	// Тест получения несуществующего платежа
	_, err = service.GetPayment("non_existent_id")
	if err == nil {
		t.Fatal("Expected error for non-existent payment, got nil")
	}
}

func TestCancelPayment(t *testing.T) {
	service := NewMockPaymentService()

	// Создаем платеж для тестирования
	request := PaymentRequest{
		OrderID:  "order_123",
		Amount:   100.00,
		Currency: "RUB",
		Method:   MethodCard,
	}

	response, _ := service.CreatePayment(request)
	paymentID := response.PaymentID

	// Тест успешной отмены платежа
	canceledPayment, err := service.CancelPayment(paymentID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if canceledPayment.Status != StatusFailed {
		t.Errorf("Expected Status %s, got %s", StatusFailed, canceledPayment.Status)
	}

	if canceledPayment.ErrorMessage == "" {
		t.Error("Expected non-empty ErrorMessage")
	}

	// Тест отмены несуществующего платежа
	_, err = service.CancelPayment("non_existent_id")
	if err == nil {
		t.Fatal("Expected error for non-existent payment, got nil")
	}

	// Тест отмены уже отмененного платежа
	_, err = service.CancelPayment(paymentID)
	if err == nil {
		t.Fatal("Expected error for already canceled payment, got nil")
	}
}

func TestRefundPayment(t *testing.T) {
	service := NewMockPaymentService()

	// Создаем платеж для тестирования
	request := PaymentRequest{
		OrderID:  "order_123",
		Amount:   100.00,
		Currency: "RUB",
		Method:   MethodCard,
	}

	response, _ := service.CreatePayment(request)
	paymentID := response.PaymentID

	// Имитируем успешное завершение платежа
	now := time.Now()
	response.Status = StatusCompleted
	response.CompletedAt = &now

	// Тест успешного возврата платежа
	refundedPayment, err := service.RefundPayment(paymentID, 50.00)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if refundedPayment.Status != StatusRefunded {
		t.Errorf("Expected Status %s, got %s", StatusRefunded, refundedPayment.Status)
	}

	// Тест возврата несуществующего платежа
	_, err = service.RefundPayment("non_existent_id", 50.00)
	if err == nil {
		t.Fatal("Expected error for non-existent payment, got nil")
	}

	// Тест возврата с суммой больше платежа
	_, err = service.RefundPayment(paymentID, 200.00)
	if err == nil {
		t.Fatal("Expected error for refund amount greater than payment amount, got nil")
	}
}

func TestCharge(t *testing.T) {
	service := NewMockPaymentService()

	// Тест успешного платежа
	result, err := service.Charge(100.00, "RUB")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == "" {
		t.Fatal("Expected non-empty result")
	}

	// Тест с некорректной суммой
	_, err = service.Charge(-100.00, "RUB")
	if err == nil {
		t.Fatal("Expected error for negative amount, got nil")
	}
}
