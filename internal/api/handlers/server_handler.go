package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/danghamo/life/internal/api/jsonrpcx"
)

// ServerHandler handles server information requests
type ServerHandler struct{}

// NewServerHandler creates a new server handler
func NewServerHandler() *ServerHandler {
	return &ServerHandler{}
}

// ServerInfoResponse represents server information
type ServerInfoResponse struct {
	Host string `json:"host"`
	Port string `json:"port"`
	URL  string `json:"url"`
}

// HandleServerInfo handles POST /api/v1/server.Info
func (h *ServerHandler) HandleServerInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	// Get server port from environment
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080" // Default port
	}

	response := ServerInfoResponse{
		Host: "localhost", // For local development
		Port: serverPort,
		URL:  fmt.Sprintf("http://localhost:%s", serverPort),
	}

	jsonrpcx.Success(w, req.ID, response)
}

// === AutoRouter Compatible Methods ===
// These methods are designed to work with the autorouter package

// Info handles server info retrieval (autorouter compatible)
func (h *ServerHandler) Info(w http.ResponseWriter, r *http.Request) {
	h.HandleServerInfo(w, r)
}