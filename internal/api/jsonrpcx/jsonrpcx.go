package jsonrpcx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      any             `json:"id,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      any           `json:"id,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// JSON-RPC 2.0 error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// ParseRequest parses JSON-RPC 2.0 request from HTTP request body
func ParseRequest(r *http.Request) (*JSONRPCRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var req JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	// Validate JSON-RPC version
	if req.JSONRPC != "2.0" {
		return nil, err
	}

	return &req, nil
}

// Success sends a successful JSON-RPC 2.0 response
func Success(w http.ResponseWriter, id any, result any) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}

	Response(w, response)
}

// WithError attaches an error to the request context for middleware processing
func WithError(r *http.Request, id any, code int, message string) {
	response := &JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}

	// Store the JSON-RPC response in request context and overwrite the request pointer
	ctx := context.WithValue(r.Context(), "jsonrpc_error", response)
	*r = *r.WithContext(ctx)
}

// ErrorAdapter interface for middleware to send error responses
type ErrorAdapter interface {
	SendError(w http.ResponseWriter, id any, code int, message string)
}

// errorAdapter is the private implementation of ErrorAdapter
type errorAdapter struct{}

// NewErrorAdapter creates a new error adapter for middleware use
func NewErrorAdapter() ErrorAdapter {
	return &errorAdapter{}
}

// SendError sends an error JSON-RPC 2.0 response (only accessible through ErrorAdapter)
func (ea *errorAdapter) SendError(w http.ResponseWriter, id any, code int, message string) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}

	Response(w, response)
}

// SetError stores a JSON-RPC error in the request context for middleware processing
func SetError(r *http.Request, id any, code int, message string) *http.Request {
	response := &JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}

	ctx := context.WithValue(r.Context(), "jsonrpc_error", response)
	return r.WithContext(ctx)
}

// Response sends a JSON-RPC 2.0 response (always HTTP 200)
func Response(w http.ResponseWriter, response JSONRPCResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC always returns HTTP 200

	// Encode response - if error occurs, it will be logged by middleware
	json.NewEncoder(w).Encode(response)
}
