package models

import "time"

type Message struct {
	ID        int       `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Content   string    `json:"content" db:"content"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type User struct {
	UserID      string    `json:"user_id" db:"user_id"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	Name        string    `json:"name" db:"name"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	LastSeen    time.Time `json:"last_seen" db:"last_seen"`
}

type ChatRequest struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

type ChatResponse struct {
	Response string `json:"response"`
	UserID   string `json:"user_id"`
}

type HistoryRequest struct {
	UserID string `json:"user_id"`
	Limit  int    `json:"limit,omitempty"`
}

type HistoryResponse struct {
	Messages []Message `json:"messages"`
	UserID   string    `json:"user_id"`
}

type WhatsAppMessage struct {
	From      string    `json:"from"`
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}

type WhatsAppContact struct {
	Profile struct {
		Name string `json:"name"`
	} `json:"profile"`
	WaID string `json:"wa_id"`
}

type WhatsAppWebhookEntry struct {
	ID      string `json:"id"`
	Changes []struct {
		Value struct {
			Messages []WhatsAppMessage `json:"messages"`
			Contacts []WhatsAppContact `json:"contacts"`
		} `json:"value"`
		Field string `json:"field"`
	} `json:"changes"`
}

type WhatsAppWebhook struct {
	Object string                 `json:"object"`
	Entry  []WhatsAppWebhookEntry `json:"entry"`
}
