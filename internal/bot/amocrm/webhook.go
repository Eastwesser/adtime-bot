package amocrm

import (
	"adtime-bot/internal/bot"
	"context"
	"encoding/json"
	"net/http"
)

type WebhookHandler struct {
	bot *bot.Bot
}

func NewWebhookHandler(bot *bot.Bot) *WebhookHandler {
	return &WebhookHandler{bot: bot}
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		EventType string          `json:"event_type"`
		Data      json.RawMessage `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch payload.EventType {
	case "lead_added":
		var lead struct {
			ID      int    `json:"id"`
			Name    string `json:"name"`
			Phone   string `json:"phone"`
			UserID  int    `json:"user_id"`
		}
		if err := json.Unmarshal(payload.Data, &lead); err == nil {
			h.bot.HandleNewLead(context.Background(), lead)
		}
	// Add other event types as needed
	}

	w.WriteHeader(http.StatusOK)
}