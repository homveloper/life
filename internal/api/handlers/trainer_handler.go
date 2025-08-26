package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/danghamo/life/internal/api/jsonrpcx"
	"github.com/danghamo/life/internal/api/middleware"
	"github.com/danghamo/life/internal/domain/shared"
	"github.com/danghamo/life/internal/domain/trainer"
	"github.com/danghamo/life/pkg/logger"
)

// TrainerHandler handles trainer-related HTTP requests with JSON-RPC 2.0 format
type TrainerHandler struct {
	logger     *logger.Logger
	repository trainer.Repository
}

// getOrCreateTrainer gets an existing trainer or creates a default one for the user
func (h *TrainerHandler) getOrCreateTrainer(ctx context.Context, userID string, defaultNickname string) (*trainer.Trainer, error) {
	trainerUserID := trainer.UserID(userID)
	
	// Try to get existing trainer
	existing, err := h.repository.GetByID(ctx, trainerUserID)
	if err != nil {
		return nil, err
	}
	
	if existing != nil {
		return existing, nil
	}
	
	// Create new trainer with default nickname
	nickname, err := trainer.NewNickname(defaultNickname)
	if err != nil {
		// If default nickname fails, use a fallback
		nickname, _ = trainer.NewNickname("Player" + userID[:8])
	}
	
	var newTrainer *trainer.Trainer
	err = h.repository.FindOneAndInsert(ctx, trainerUserID, func() (*trainer.Trainer, error) {
		return trainer.NewTrainer(trainerUserID, nickname)
	})
	if err != nil {
		return nil, err
	}
	
	// Get the created trainer
	newTrainer, err = h.repository.GetByID(ctx, trainerUserID)
	if err != nil {
		return nil, err
	}
	
	h.logger.Info("Auto-created trainer for new user", 
		zap.String("userId", userID),
		zap.String("nickname", nickname.Value()))
	
	return newTrainer, nil
}

// NewTrainerHandler creates a new trainer handler
func NewTrainerHandler(logger *logger.Logger, repository trainer.Repository) *TrainerHandler {
	return &TrainerHandler{
		logger:     logger.WithComponent("trainer-handler"),
		repository: repository,
	}
}

// Request parameter structures
type CreateTrainerParams struct {
	Nickname string `json:"nickname"`
}

type GetTrainerParams struct {
	// No ID needed - we get it from JWT context
}

type MoveTrainerParams struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// HandleCreate handles POST /api/v1/trainer.Create
func (h *TrainerHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	// Get user info from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		jsonrpcx.WithError(r, nil, jsonrpcx.InvalidRequest, "User not authenticated")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	var params CreateTrainerParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	// Create trainer domain entity
	nickname, err := trainer.NewNickname(params.Nickname)
	if err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, err.Error())
		return
	}

	var createdTrainer *trainer.Trainer
	trainerUserID := trainer.UserID(userID)

	err = h.repository.FindOneAndInsert(r.Context(), trainerUserID, func() (*trainer.Trainer, error) {
		return trainer.NewTrainer(trainerUserID, nickname)
	})
	if err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to create trainer")
		return
	}

	// Get created trainer for response
	createdTrainer, err = h.repository.GetByID(r.Context(), trainerUserID)
	if err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to retrieve created trainer")
		return
	}

	result := createdTrainer

	h.logger.Info("Trainer created successfully", 
		zap.String("userId", userID),
		zap.String("nickname", createdTrainer.Nickname.Value()))

	jsonrpcx.Success(w, req.ID, result)
}

// HandleGet handles POST /api/v1/trainer.Get
func (h *TrainerHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	// Get user info from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		jsonrpcx.WithError(r, nil, jsonrpcx.InvalidRequest, "User not authenticated")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	var params GetTrainerParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	// Get trainer from repository using UserID from JWT (auto-create if not exists)
	trainerEntity, err := h.getOrCreateTrainer(r.Context(), userID, "NewPlayer")
	if err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to retrieve trainer")
		return
	}

	result := trainerEntity

	jsonrpcx.Success(w, req.ID, result)
}

// HandleMove handles POST /api/v1/trainer.Move
func (h *TrainerHandler) HandleMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	// Get user info from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		jsonrpcx.WithError(r, nil, jsonrpcx.InvalidRequest, "User not authenticated")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	var params MoveTrainerParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	// Movement validation
	if err := h.validateMovement(params.X, params.Y); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, err.Error())
		return
	}

	// Get or create trainer first
	_, err = h.getOrCreateTrainer(r.Context(), userID, "NewPlayer")
	if err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to get trainer")
		return
	}

	// Update trainer position using repository
	newPosition := shared.NewPosition(params.X, params.Y)
	var updatedTrainer *trainer.Trainer
	trainerUserID := trainer.UserID(userID)

	err = h.repository.FindOneAndUpdate(r.Context(), trainerUserID, func(t *trainer.Trainer) (*trainer.Trainer, error) {
		if t == nil {
			return nil, fmt.Errorf("trainer not found")
		}

		// Move the trainer
		if err := t.MoveTo(newPosition); err != nil {
			return nil, err
		}

		updatedTrainer = t
		return t, nil
	})

	if err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, fmt.Sprintf("Failed to move trainer: %v", err))
		return
	}

	result := updatedTrainer

	h.logger.Info("Trainer movement",
		zap.String("userId", userID),
		zap.Int("newX", updatedTrainer.Position.X),
		zap.Int("newY", updatedTrainer.Position.Y),
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)))

	jsonrpcx.Success(w, req.ID, result)
}

// HandleList handles POST /api/v1/trainer.List
func (h *TrainerHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	// TODO: Implement actual trainer listing using query handler
	// For now, return a mock response
	result := map[string]any{
		"trainers": []map[string]any{
			{
				"id":       "trainer_1",
				"nickname": "Player1",
				"level":    3,
				"position": map[string]int{"x": 10, "y": 5},
			},
			{
				"id":       "trainer_2",
				"nickname": "Player2",
				"level":    7,
				"position": map[string]int{"x": 20, "y": 15},
			},
		},
		"total": 2,
	}

	jsonrpcx.Success(w, req.ID, result)
}

// HandleStatus handles POST /api/v1/trainer.Status
func (h *TrainerHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	// Get user info from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		jsonrpcx.WithError(r, nil, jsonrpcx.InvalidRequest, "User not authenticated")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	var params GetTrainerParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	// Get trainer status from repository using UserID from JWT (auto-create if not exists)
	trainerEntity, err := h.getOrCreateTrainer(r.Context(), userID, "NewPlayer")
	if err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to retrieve trainer status")
		return
	}

	result := trainerEntity

	jsonrpcx.Success(w, req.ID, result)
}

// validateMovement validates trainer movement coordinates
func (h *TrainerHandler) validateMovement(x, y int) error {
	// Map boundaries validation (assuming 30x20 map from world handler)
	const maxX, maxY = 30, 20

	if x < 0 || x >= maxX {
		return fmt.Errorf("invalid X coordinate: %d (must be 0-%d)", x, maxX-1)
	}

	if y < 0 || y >= maxY {
		return fmt.Errorf("invalid Y coordinate: %d (must be 0-%d)", y, maxY-1)
	}

	// TODO: Add obstacle checking logic from world data
	// For now, just check for basic water tiles at specific coordinates
	waterTiles := []struct{ x, y int }{
		{2, 0}, {0, 2}, // From world handler mock data
	}

	for _, tile := range waterTiles {
		if x == tile.x && y == tile.y {
			return fmt.Errorf("cannot move to water tile at (%d, %d)", x, y)
		}
	}

	return nil
}
