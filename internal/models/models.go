package models

import "time"

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
)

type Customer struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type Parcel struct {
	ID        int       `json:"id"`
	ClientID  int       `json:"client_id"`
	Address   string    `json:"address"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Delivery struct {
	ID          int       `json:"id"`
	CourierID   int       `json:"courier_id"`
	ParcelID    int       `json:"parcel_id"`
	Status      string    `json:"status"`
	AssignedAt  time.Time `json:"assigned_at"`
	DeliveredAt time.Time `json:"delivered_at"`
}
