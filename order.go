package main

import (
	"math/rand"
	"time"
)

type Order struct {
	OrderID    string `json:"orderID"`
	DeliveryID string `json:"deliveryID"`
	Status     string `json:"status"`
}

func GenerateID(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

func GenerateOrder() Order {
	var order Order = Order{
		OrderID:    GenerateID(8),
		DeliveryID: GenerateID(6),
		Status:     "pending",
	}
	return order
}
