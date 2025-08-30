package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/danghamo/life/internal/api/jsonrpcx"
	"github.com/danghamo/life/internal/domain/shared"
	"github.com/danghamo/life/pkg/logger"
)

// AnimalHandler handles animal-related HTTP requests with JSON-RPC 2.0 format
type AnimalHandler struct {
	logger *logger.Logger
}

// NewAnimalHandler creates a new animal handler
func NewAnimalHandler(logger *logger.Logger) *AnimalHandler {
	return &AnimalHandler{
		logger: logger.WithComponent("animal-handler"),
	}
}

// Request parameter structures
type SpawnAnimalParams struct {
	Type  string `json:"type"`
	Level int    `json:"level"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
}

type GetAnimalParams struct {
	ID string `json:"id"`
}

type CaptureAnimalParams struct {
	AnimalID  string `json:"animalId"`
	TrainerID string `json:"trainerId"`
}

// HandleSpawn handles POST /api/v1/animal.Spawn
func (h *AnimalHandler) HandleSpawn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	var params SpawnAnimalParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	// TODO: Implement actual animal spawning using command handler
	// For now, return a mock response
	result := map[string]any{
		"id":       "animal_" + shared.NewID().String(),
		"type":     params.Type,
		"level":    params.Level,
		"position": map[string]int{"x": params.X, "y": params.Y},
		"state":    "wild",
		"status":   "spawned",
	}

	jsonrpcx.Success(w, req.ID, result)
}

// HandleGet handles POST /api/v1/animal.Get
func (h *AnimalHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	var params GetAnimalParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	// TODO: Implement actual animal retrieval using query handler
	// For now, return a mock response
	result := map[string]any{
		"id":       params.ID,
		"type":     "lion",
		"level":    3,
		"position": map[string]int{"x": 12, "y": 8},
		"state":    "wild",
		"stats": map[string]int{
			"hp":  90,
			"atk": 18,
			"def": 12,
			"spd": 15,
			"as":  14,
		},
	}

	jsonrpcx.Success(w, req.ID, result)
}

// HandleCapture handles POST /api/v1/animal.Capture
func (h *AnimalHandler) HandleCapture(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	var params CaptureAnimalParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	// TODO: Implement actual animal capture using command handler
	// For now, return a mock response
	result := map[string]any{
		"animalId":  params.AnimalID,
		"trainerId": params.TrainerID,
		"status":    "captured",
		"success":   true,
		"message":   "Animal captured successfully!",
	}

	jsonrpcx.Success(w, req.ID, result)
}

// HandleList handles POST /api/v1/animal.List
func (h *AnimalHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	// TODO: Implement actual animal listing using query handler
	// For now, return a mock response
	result := map[string]any{
		"animals": []map[string]any{
			{
				"id":       "animal_1",
				"type":     "lion",
				"level":    5,
				"position": map[string]int{"x": 15, "y": 10},
				"state":    "wild",
			},
			{
				"id":       "animal_2",
				"type":     "elephant",
				"level":    8,
				"position": map[string]int{"x": 25, "y": 18},
				"state":    "captured",
			},
		},
		"total": 2,
	}

	jsonrpcx.Success(w, req.ID, result)
}

// === AutoRouter Compatible Methods ===
// These methods are designed to work with the autorouter package

// Spawn handles animal spawning (autorouter compatible)
func (h *AnimalHandler) Spawn(w http.ResponseWriter, r *http.Request) {
	h.HandleSpawn(w, r)
}

// Get handles animal retrieval (autorouter compatible)
func (h *AnimalHandler) Get(w http.ResponseWriter, r *http.Request) {
	h.HandleGet(w, r)
}

// Capture handles animal capture (autorouter compatible)
func (h *AnimalHandler) Capture(w http.ResponseWriter, r *http.Request) {
	h.HandleCapture(w, r)
}

// List handles animal listing (autorouter compatible)
func (h *AnimalHandler) List(w http.ResponseWriter, r *http.Request) {
	h.HandleList(w, r)
}
