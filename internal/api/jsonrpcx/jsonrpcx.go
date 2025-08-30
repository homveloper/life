package jsonrpcx

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// RequestT represents a typed JSON-RPC 2.0 request
type RequestT[T any] struct {
	JSONRPC string `json:"jsonrpc" example:"2.0"`
	Method  string `json:"method"`
	Params  T      `json:"params,omitempty"`
	ID      any    `json:"id,omitempty"`
}

// ResponseT represents a typed JSON-RPC 2.0 response  
type ResponseT[T any] struct {
	JSONRPC string        `json:"jsonrpc" example:"2.0"`
	Result  T             `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      any           `json:"id,omitempty"`
}

// JsonRpcNotification represents a JSON-RPC 2.0 notification (no ID, no response expected)
type JsonRpcNotification struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// Request represents a JSON-RPC 2.0 request
type Request struct {
	JSONRPC string          `json:"jsonrpc" example:"2.0"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      any             `json:"id,omitempty"`
}

// Response represents a JSON-RPC 2.0 response
type Response struct {
	JSONRPC string        `json:"jsonrpc" example:"2.0"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      any           `json:"id,omitempty"`
}

// ErrorResponse represents a JSON-RPC 2.0 error response for Swagger
type ErrorResponse struct {
	JSONRPC string        `json:"jsonrpc" example:"2.0"`
	Error   *JSONRPCError `json:"error"`
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
func ParseRequest(r *http.Request) (*Request, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var req Request
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
	response := Response{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}

	Write(w, response)
}

// WithError attaches an error to the request context for middleware processing
func WithError(r *http.Request, id any, code int, message string) {
	response := &Response{
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
	response := Response{
		JSONRPC: "2.0",
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}

	Write(w, response)
}

// SetError stores a JSON-RPC error in the request context for middleware processing
func SetError(r *http.Request, id any, code int, message string) *http.Request {
	response := &Response{
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

// Write sends a JSON-RPC 2.0 response (always HTTP 200)
func Write(w http.ResponseWriter, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC always returns HTTP 200

	// Encode response - if error occurs, it will be logged by middleware
	json.NewEncoder(w).Encode(response)
}
