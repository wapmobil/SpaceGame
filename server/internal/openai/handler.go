package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type ChatHandler struct {
	Client *Client
}

func NewChatHandler() *ChatHandler {
	return &ChatHandler{
		Client: New(),
	}
}

type ChatHandlerRequest struct {
	Prompt string `json:"prompt"`
}

type ChatHandlerResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatHandlerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ChatHandlerResponse{Error: "invalid request body"})
		return
	}

	if req.Prompt == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ChatHandlerResponse{Error: "prompt is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	result, err := h.Client.Chat(ctx, req.Prompt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ChatHandlerResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatHandlerResponse{Response: result})
}
