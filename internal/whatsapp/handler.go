package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sinhaparth5/whatstyle-mcp/configs"
	"github.com/sinhaparth5/whatstyle-mcp/internal/models"
)

type Handler struct {
	config     *configs.Config
	httpClient *http.Client
}

type SendMessageRequest struct {
	MessagingProduct string `json:"messaging_product"`
	To               string `json:"to"`
	Type             string `json:"type"`
	Text             struct {
		Body string `json:"body"`
	} `json:"text"`
}

type SendMessageResponse struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

func NewHandler(config *configs.Config) *Handler {
	return &Handler{
		config:     config,
		httpClient: &http.Client{},
	}
}

func (h *Handler) VerifyWebhook(w http.ResponseWriter, r *http.Request) {
	// Verify webhook for WhatsApp Business API
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	if mode == "subscribe" && token == h.config.WhatsAppVerifyToken {
		log.Printf("Webhook verified successfully")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(challenge))
		return
	}

	log.Printf("Webhook verification failed")
	w.WriteHeader(http.StatusForbidden)
}

func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var webhook models.WhatsAppWebhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		log.Printf("Error decoding webhook: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Process webhook entries
	for _, entry := range webhook.Entry {
		for _, change := range entry.Changes {
			if change.Field == "messages" {
				h.processMessages(change.Value.Messages, change.Value.Contacts)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) processMessages(messages []models.WhatsAppMessage, contacts []models.WhatsAppContact) {
	for _, message := range messages {
		// Get contact info
		var contactName string
		for _, contact := range contacts {
			if contact.WaID == message.From {
				contactName = contact.Profile.Name
				break
			}
		}

		log.Printf("Received message from %s (%s): %s", message.From, contactName, message.Text)

		// Here you would integrate with your MCP chat handler
		// For now, just log the message
		// TODO: Integrate with MCP chat tool
	}
}

func (h *Handler) SendMessage(to, message string) error {
	if h.config.WhatsAppAccessToken == "" {
		return fmt.Errorf("WhatsApp access token not configured")
	}

	url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", h.config.WhatsAppPhoneNumberID)

	reqBody := SendMessageRequest{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             "text",
	}
	reqBody.Text.Body = message

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.config.WhatsAppAccessToken)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WhatsApp API returned status %d", resp.StatusCode)
	}

	var response SendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("Message sent successfully to %s, ID: %s", to, response.Messages[0].ID)
	return nil
}
