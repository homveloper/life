package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/danghamo/life/internal/api/jsonrpcx"
	"github.com/danghamo/life/pkg/logger"
)

// WorldHandler handles world-related HTTP requests with JSON-RPC 2.0 format
type WorldHandler struct {
	logger *logger.Logger
}

// NewWorldHandler creates a new world handler
func NewWorldHandler(logger *logger.Logger) *WorldHandler {
	return &WorldHandler{
		logger: logger.WithComponent("world-handler"),
	}
}

// Request parameter structures
type GetWorldParams struct {
	ID string `json:"id"`
}

// HandleGet handles POST /api/v1/world.Get
func (h *WorldHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	var params GetWorldParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	// TODO: Implement actual world retrieval using query handler
	// For now, return a mock response
	result := map[string]any{
		"id":     params.ID,
		"name":   "Test World",
		"width":  30,
		"height": 20,
		"status": "active",
		"terrain": [][]string{
			{"grass", "grass", "water", "grass"},
			{"grass", "rock", "grass", "grass"},
			{"water", "grass", "grass", "rock"},
		},
		"spawn_points": []map[string]int{
			{"x": 5, "y": 5, "type": 1},
			{"x": 15, "y": 10, "type": 2},
			{"x": 25, "y": 15, "type": 3},
		},
	}

	jsonrpcx.Success(w, req.ID, result)
}
