package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestChat_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []Choice{
				{
					Message: ChatMessage{
						Role:    "assistant",
						Content: "Hello from mock server!",
					},
				},
			},
		})
	}))
	defer mockServer.Close()

	client := &Client{
		BaseURL: mockServer.URL,
		Model:   "test-model",
		HTTP:    mockServer.Client(),
	}

	ctx := context.Background()
	result, err := client.Chat(ctx, "test prompt")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != "Hello from mock server!" {
		t.Fatalf("expected 'Hello from mock server!', got: %s", result)
	}
}

func TestChat_EmptyResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []Choice{},
		})
	}))
	defer mockServer.Close()

	client := &Client{
		BaseURL: mockServer.URL,
		Model:   "test-model",
		HTTP:    mockServer.Client(),
	}

	ctx := context.Background()
	_, err := client.Chat(ctx, "test prompt")
	if err == nil {
		t.Fatal("expected error for empty response, got nil")
	}
}

func TestChat_ServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer mockServer.Close()

	client := &Client{
		BaseURL: mockServer.URL,
		Model:   "test-model",
		HTTP:    mockServer.Client(),
	}

	ctx := context.Background()
	_, err := client.Chat(ctx, "test prompt")
	if err == nil {
		t.Fatal("expected error for server error, got nil")
	}
}

func TestChat_ContextCanceled(t *testing.T) {
	client := &Client{
		BaseURL: "http://nonexistent:99999",
		Model:   "test-model",
		HTTP: &http.Client{
			Timeout: 100 * time.Millisecond,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Chat(ctx, "test prompt")
	if err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}
}
