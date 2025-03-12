package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем подключения с любых источников (в продакшене лучше настроить более строго)
	},
}

// Структура для хранения клиентов WebSocket
type WebSocketManager struct {
	clients    map[*websocket.Conn]bool
	clientsMu  sync.RWMutex
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

// Создаем новый менеджер WebSocket
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Запускаем обработку сообщений в отдельной горутине
func (manager *WebSocketManager) Run() {
	for {
		select {
		case client := <-manager.register:
			manager.clientsMu.Lock()
			manager.clients[client] = true
			manager.clientsMu.Unlock()
		case client := <-manager.unregister:
			manager.clientsMu.Lock()
			if _, ok := manager.clients[client]; ok {
				delete(manager.clients, client)
				client.Close()
			}
			manager.clientsMu.Unlock()
		case message := <-manager.broadcast:
			manager.clientsMu.RLock()
			for client := range manager.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					log.Printf("Ошибка при отправке сообщения: %v", err)
					client.Close()
					delete(manager.clients, client)
				}
			}
			manager.clientsMu.RUnlock()
		}
	}
}

// Обработчик WebSocket соединений
func (manager *WebSocketManager) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка при установке WebSocket соединения: %v", err)
		return
	}

	// Регистрируем нового клиента
	manager.register <- conn

	// Обрабатываем сообщения от клиента
	go func() {
		defer func() {
			manager.unregister <- conn
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Ошибка при чтении сообщения: %v", err)
				}
				break
			}
		}
	}()
}

// GetActiveConnectionsCount возвращает количество активных WebSocket соединений
func (manager *WebSocketManager) GetActiveConnectionsCount() int {
	manager.clientsMu.RLock()
	defer manager.clientsMu.RUnlock()
	return len(manager.clients)
}

// Структура для сообщения об обновлении статуса заказа
type OrderStatusUpdate struct {
	Type      string `json:"type"`
	OrderID   string `json:"orderId"`
	NewStatus string `json:"newStatus"`
	Timestamp int64  `json:"timestamp"`
}

// Отправляет обновление статуса заказа всем подключенным клиентам
func (manager *WebSocketManager) BroadcastOrderStatusUpdate(orderID, newStatus string) {
	update := OrderStatusUpdate{
		Type:      "ORDER_STATUS_UPDATE",
		OrderID:   orderID,
		NewStatus: newStatus,
		Timestamp: GetCurrentTimestamp(),
	}

	jsonData, err := json.Marshal(update)
	if err != nil {
		log.Printf("Ошибка при сериализации обновления статуса: %v", err)
		return
	}

	manager.broadcast <- jsonData
}

// Получение текущего времени в миллисекундах
func GetCurrentTimestamp() int64 {
	return int64(time.Now().UnixNano() / int64(time.Millisecond))
}
