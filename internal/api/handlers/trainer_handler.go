package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	jsonpatch "github.com/evanphx/json-patch/v5"
	"go.uber.org/zap"

	"github.com/danghamo/life/internal/api/jsonrpcx"
	"github.com/danghamo/life/internal/api/middleware"
	cqrscommands "github.com/danghamo/life/internal/cqrs"
	"github.com/danghamo/life/internal/domain/shared"
	"github.com/danghamo/life/internal/domain/trainer"
	"github.com/danghamo/life/pkg/logger"
)

// MovementBroadcaster interface for broadcasting moving trainer positions
type MovementBroadcaster interface {
	AddMovingTrainer(userID, displayName, color string) // displayName can be userID or nickname
	RemoveMovingTrainer(userID string)
	UpdateTrainerActivity(userID string)
	GetCurrentOnlineTrainers(ctx context.Context) []cqrscommands.TrainerMovedEvent
}

// TrainerHandler handles trainer-related HTTP requests with JSON-RPC 2.0 format
type TrainerHandler struct {
	logger              *logger.Logger
	repository          trainer.Repository
	eventBus            *cqrs.EventBus
	movementBroadcaster MovementBroadcaster
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
	nickname := defaultNickname
	if err := trainer.ValidateNickname(nickname); err != nil {
		// If default nickname fails, use a fallback
		nickname = "Player" + userID[:8]
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
		zap.String("nickname", nickname))

	return newTrainer, nil
}

// NewTrainerHandler creates a new trainer handler
func NewTrainerHandler(logger *logger.Logger, repository trainer.Repository, eventBus *cqrs.EventBus, movementBroadcaster MovementBroadcaster) *TrainerHandler {
	return &TrainerHandler{
		logger:              logger.WithComponent("trainer-handler"),
		repository:          repository,
		eventBus:            eventBus,
		movementBroadcaster: movementBroadcaster,
	}
}

// Request parameter structures
type CreateTrainerRequest struct {
	Nickname string `json:"nickname"`
}

type GetTrainerRequest struct {
	// No ID needed - we get it from JWT context
}

type MoveTrainerRequest struct {
	DirectionX float64 `json:"direction_x"` // -1, 0, or 1
	DirectionY float64 `json:"direction_y"` // -1, 0, or 1
	Action     string  `json:"action"`      // "start" or "stop"
}

type ListTrainerRequest struct {
	OnlineOnly bool `json:"online_only,omitempty"` // Filter to show only currently online trainers
}

type FetchPositionRequest struct {
	// No ID needed - we get it from JWT context
}

type FetchPositionResponse struct {
	Position shared.Position       `json:"position"`
	Movement trainer.MovementState `json:"movement"`
}

// Response structures for Swagger documentation
type CreateTrainerResponse = trainer.Trainer
type GetTrainerResponse = trainer.Trainer
type MoveTrainerResponse struct {
	Changes              map[string]interface{} `json:"changes"`
	NextRequestAllowedAt int64                  `json:"next_request_allowed_at"` // Unix timestamp in milliseconds
}
type StatusTrainerResponse = trainer.Trainer

type ListTrainerResponse struct {
	Trainers []TrainerSummary `json:"trainers"`
	Total    int              `json:"total"`
}

type TrainerSummary struct {
	ID       string         `json:"id"`
	Nickname string         `json:"nickname"`
	Color    string         `json:"color"`
	Level    int            `json:"level"`
	Position map[string]int `json:"position"`
}

// HandleCreate handles POST /api/v1/trainer.Create
// @Summary Create a new trainer
// @Description Create a new trainer with nickname for the authenticated user
// @Tags trainer
// @Accept json
// @Produce json
// @Param request body jsonrpcx.RequestT[CreateTrainerRequest] true "JSON-RPC request with CreateTrainerRequest params"
// @Success 200 {object} jsonrpcx.ResponseT[CreateTrainerResponse] "Created trainer information"
// @Failure 400 {object} jsonrpcx.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} jsonrpcx.ErrorResponse "Authentication required"
// @Failure 500 {object} jsonrpcx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/trainer.Create [post]
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

	var params CreateTrainerRequest
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	// Create trainer domain entity
	nickname := params.Nickname
	if err := trainer.ValidateNickname(nickname); err != nil {
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
		zap.String("nickname", createdTrainer.Nickname))

	jsonrpcx.Success(w, req.ID, result)
}

// HandleGet handles POST /api/v1/trainer.Get
// @Summary Get trainer information
// @Description Get trainer information for the authenticated user (auto-creates if not exists)
// @Tags trainer
// @Accept json
// @Produce json
// @Param request body jsonrpcx.RequestT[GetTrainerRequest] true "JSON-RPC request with GetTrainerRequest params"
// @Success 200 {object} jsonrpcx.ResponseT[GetTrainerResponse] "Trainer information"
// @Failure 400 {object} jsonrpcx.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} jsonrpcx.ErrorResponse "Authentication required"
// @Failure 500 {object} jsonrpcx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/trainer.Get [post]
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

	var params GetTrainerRequest
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
// @Summary Move trainer to new position
// @Description Move trainer to a new position on the game map with coordinate validation
// @Tags trainer
// @Accept json
// @Produce json
// @Param request body jsonrpcx.RequestT[MoveTrainerRequest] true "JSON-RPC request with MoveTrainerRequest params"
// @Success 200 {object} jsonrpcx.ResponseT[MoveTrainerResponse] "Updated trainer with new position"
// @Failure 400 {object} jsonrpcx.ErrorResponse "Invalid request parameters or invalid coordinates"
// @Failure 401 {object} jsonrpcx.ErrorResponse "Authentication required"
// @Failure 500 {object} jsonrpcx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/trainer.Move [post]
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

	var params MoveTrainerRequest
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	// Validate direction
	if err := h.validateDirection(params.DirectionX, params.DirectionY); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, err.Error())
		return
	}

	// Get or create trainer first
	_, err = h.getOrCreateTrainer(r.Context(), userID, "NewPlayer")
	if err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to get trainer")
		return
	}

	// Get original trainer state for comparison
	originalTrainer, err := h.repository.GetByID(r.Context(), trainer.UserID(userID))
	if err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to get original trainer state")
		return
	}
	

	// Handle movement command
	var updatedTrainer *trainer.Trainer
	trainerUserID := trainer.UserID(userID)

	err = h.repository.FindOneAndUpdate(r.Context(), trainerUserID, func(t *trainer.Trainer) (*trainer.Trainer, error) {
		if t == nil {
			return nil, fmt.Errorf("trainer not found")
		}

		// Update position from current movement before new command
		t.UpdatePositionFromMovement()

		// Handle movement action
		if params.Action == "start" {
			if err := t.StartMovement(params.DirectionX, params.DirectionY); err != nil {
				return nil, err
			}
		} else if params.Action == "stop" {
			if err := t.StopMovement(); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("invalid action: %s (must be 'start' or 'stop')", params.Action)
		}

		updatedTrainer = t
		return t, nil
	})

	if err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, fmt.Sprintf("Failed to move trainer: %v", err))
		return
	}

	// Create JSON merge patch with only changed fields  
	var changes map[string]interface{}
	if originalTrainer != nil {
		changes, err = h.createTrainerChanges(originalTrainer, updatedTrainer)
		if err != nil {
			h.logger.Warn("Failed to create changes patch", zap.Error(err))
			// Continue with empty changes rather than failing the request
			changes = make(map[string]interface{})
		}
	} else {
		// No original trainer to compare with, return empty changes
		changes = make(map[string]interface{})
	}

	// Publish domain event for SSE broadcasting
	requestID := fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())
	var event interface{}

	if params.Action == "start" {
		// Add to movement broadcaster for periodic position updates
		h.movementBroadcaster.AddMovingTrainer(userID, userID, updatedTrainer.Color)
		
		event = &cqrscommands.TrainerMovedEvent{
			UserID:    userID,
			Nickname:  updatedTrainer.Nickname,
			Color:     updatedTrainer.Color,
			Position:  updatedTrainer.Position,
			Movement:  updatedTrainer.Movement,
			Timestamp: time.Now(),
			RequestID: requestID,
			Changes:   changes,
		}
	} else {
		// Remove from movement broadcaster when stopped
		h.movementBroadcaster.RemoveMovingTrainer(userID)
		
		event = &cqrscommands.TrainerStoppedEvent{
			UserID:    userID,
			Nickname:  updatedTrainer.Nickname,
			Color:     updatedTrainer.Color,
			Position:  updatedTrainer.Position,
			Movement:  updatedTrainer.Movement,
			Timestamp: time.Now(),
			RequestID: requestID,
			Changes:   changes,
		}
	}

	// Publish event for SSE broadcasting
	if err := h.eventBus.Publish(r.Context(), event); err != nil {
		h.logger.Error("Failed to publish trainer movement event",
			zap.Error(err),
			zap.String("userId", userID),
			zap.String("action", params.Action))
		// Don't fail the request if event publishing fails
	}

	// Calculate next request allowed timestamp (100ms debounce)
	const debounceMillis = 100
	nextAllowedAt := time.Now().Add(debounceMillis * time.Millisecond).UnixMilli()

	result := MoveTrainerResponse{
		Changes:              changes,
		NextRequestAllowedAt: nextAllowedAt,
	}

	h.logger.Info("Trainer movement command",
		zap.String("userId", userID),
		zap.String("action", params.Action),
		zap.Float64("directionX", params.DirectionX),
		zap.Float64("directionY", params.DirectionY),
		zap.Float64("currentX", updatedTrainer.Position.X),
		zap.Float64("currentY", updatedTrainer.Position.Y),
		zap.Bool("isMoving", updatedTrainer.Movement.IsMoving))

	jsonrpcx.Success(w, req.ID, result)
}

// HandleList handles POST /api/v1/trainer.List
// @Summary List all trainers
// @Description Get a list of all trainers in the game world (currently returns mock data)
// @Tags trainer
// @Accept json
// @Produce json
// @Param request body jsonrpcx.RequestT[ListTrainerRequest] true "JSON-RPC request with ListTrainerRequest params"
// @Success 200 {object} jsonrpcx.ResponseT[ListTrainerResponse] "List of trainers"
// @Failure 400 {object} jsonrpcx.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} jsonrpcx.ErrorResponse "Authentication required"
// @Failure 500 {object} jsonrpcx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/trainer.List [post]
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

	// Parse request parameters
	var params ListTrainerRequest
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid request parameters")
		return
	}

	// Get current user ID to exclude from list
	currentUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidRequest, "User not authenticated")
		return
	}

	var trainerSummaries []TrainerSummary

	if params.OnlineOnly {
		// Get only currently online trainers from movement broadcaster
		h.logger.Debug("Filtering for online trainers only")
		onlineEvents := h.movementBroadcaster.GetCurrentOnlineTrainers(r.Context())
		
		for _, event := range onlineEvents {
			if event.UserID != currentUserID {
				trainerSummaries = append(trainerSummaries, TrainerSummary{
					ID:       event.UserID,
					Nickname: event.Nickname,
					Color:    event.Color,
					Level:    1, // TODO: Get actual level from repository if needed
					Position: map[string]int{
						"x": int(event.Position.X),
						"y": int(event.Position.Y),
					},
				})
			}
		}
	} else {
		// Get all trainers from repository (original behavior)
		trainers, err := h.repository.GetAll(r.Context())
		if err != nil {
			h.logger.Error("Failed to list trainers", zap.Error(err))
			jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to retrieve trainers")
			return
		}

		// Convert to response format, excluding current user
		for _, t := range trainers {
			if string(t.ID) != currentUserID {
				// Update position from movement before returning
				t.UpdatePositionFromMovement()

				trainerSummaries = append(trainerSummaries, TrainerSummary{
					ID:       string(t.ID),
					Nickname: t.Nickname,
					Color:    t.Color,
					Level:    t.Level.Value(),
					Position: map[string]int{
						"x": int(t.Position.X),
						"y": int(t.Position.Y),
					},
				})
			}
		}
	}

	result := ListTrainerResponse{
		Trainers: trainerSummaries,
		Total:    len(trainerSummaries),
	}

	jsonrpcx.Success(w, req.ID, result)
}

// HandleStatus handles POST /api/v1/trainer.Status
// @Summary Get trainer status
// @Description Get detailed status information for the authenticated trainer
// @Tags trainer
// @Accept json
// @Produce json
// @Param request body jsonrpcx.RequestT[GetTrainerRequest] true "JSON-RPC request with GetTrainerRequest params"
// @Success 200 {object} jsonrpcx.ResponseT[StatusTrainerResponse] "Trainer status information"
// @Failure 400 {object} jsonrpcx.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} jsonrpcx.ErrorResponse "Authentication required"
// @Failure 500 {object} jsonrpcx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/trainer.Status [post]
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

	var params GetTrainerRequest
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

// HandleFetchPosition handles POST /api/v1/trainer.FetchPosition
// @Summary Fetch trainer position and movement state
// @Description Get current position and movement state for synchronization fallback
// @Tags trainer
// @Accept json
// @Produce json
// @Param request body jsonrpcx.RequestT[FetchPositionRequest] true "JSON-RPC request with FetchPositionRequest params"
// @Success 200 {object} jsonrpcx.ResponseT[FetchPositionResponse] "Current position and movement state"
// @Failure 400 {object} jsonrpcx.ErrorResponse "Invalid request parameters"
// @Failure 401 {object} jsonrpcx.ErrorResponse "Authentication required"
// @Failure 500 {object} jsonrpcx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /api/v1/trainer.FetchPosition [post]
func (h *TrainerHandler) HandleFetchPosition(w http.ResponseWriter, r *http.Request) {
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

	var params FetchPositionRequest
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	// Get trainer and update position from movement
	trainerUserID := trainer.UserID(userID)
	var currentTrainer *trainer.Trainer

	err = h.repository.FindOneAndUpdate(r.Context(), trainerUserID, func(t *trainer.Trainer) (*trainer.Trainer, error) {
		if t == nil {
			return nil, fmt.Errorf("trainer not found")
		}

		// Update position from current movement state
		t.UpdatePositionFromMovement()
		currentTrainer = t
		return t, nil
	})

	if err != nil {
		// Try to get trainer without update (fallback for read-only access)
		currentTrainer, err = h.getOrCreateTrainer(r.Context(), userID, "NewPlayer")
		if err != nil {
			jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to fetch trainer position")
			return
		}
		// Update position calculation for response
		currentTrainer.UpdatePositionFromMovement()
	}

	result := FetchPositionResponse{
		Position: currentTrainer.Position,
		Movement: currentTrainer.Movement,
	}

	h.logger.Debug("Fetched trainer position",
		zap.String("userId", userID),
		zap.Float64("x", currentTrainer.Position.X),
		zap.Float64("y", currentTrainer.Position.Y),
		zap.Bool("isMoving", currentTrainer.Movement.IsMoving))

	jsonrpcx.Success(w, req.ID, result)
}

// validateDirection validates movement direction values
func (h *TrainerHandler) validateDirection(dirX, dirY float64) error {
	// Direction values must be -1, 0, or 1
	validValues := []float64{-1, 0, 1}

	isValidX := false
	for _, v := range validValues {
		if dirX == v {
			isValidX = true
			break
		}
	}

	isValidY := false
	for _, v := range validValues {
		if dirY == v {
			isValidY = true
			break
		}
	}

	if !isValidX {
		return fmt.Errorf("invalid direction_x: %f (must be -1, 0, or 1)", dirX)
	}

	if !isValidY {
		return fmt.Errorf("invalid direction_y: %f (must be -1, 0, or 1)", dirY)
	}

	return nil
}

// validateMovement validates trainer movement coordinates
func (h *TrainerHandler) validateMovement(x, y float64) error {
	// Map boundaries validation (assuming 30x20 map from world handler)
	const maxX, maxY = 30, 20

	if x < 0 || x >= maxX {
		return fmt.Errorf("invalid X coordinate: %f (must be 0-%d)", x, maxX-1)
	}

	if y < 0 || y >= maxY {
		return fmt.Errorf("invalid Y coordinate: %f (must be 0-%d)", y, maxY-1)
	}

	// TODO: Add obstacle checking logic from world data
	// For now, just check for basic water tiles at specific coordinates
	waterTiles := []struct{ x, y float64 }{
		{2, 0}, {0, 2}, // From world handler mock data
	}

	for _, tile := range waterTiles {
		if x == tile.x && y == tile.y {
			return fmt.Errorf("cannot move to water tile at (%f, %f)", x, y)
		}
	}

	return nil
}

// createTrainerChanges creates a JSON merge patch containing only changed fields
func (h *TrainerHandler) createTrainerChanges(original, updated *trainer.Trainer) (map[string]interface{}, error) {
	if original == nil || updated == nil {
		return nil, fmt.Errorf("original or updated trainer is nil")
	}

	// Marshal both trainers to JSON
	originalJSON, err := json.Marshal(original)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal original trainer: %w", err)
	}

	updatedJSON, err := json.Marshal(updated)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated trainer: %w", err)
	}

	// Create merge patch using json-patch library
	mergePatch, err := jsonpatch.CreateMergePatch(originalJSON, updatedJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create merge patch: %w", err)
	}

	// Unmarshal merge patch to map
	var changes map[string]interface{}
	if err := json.Unmarshal(mergePatch, &changes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal merge patch: %w", err)
	}

	return changes, nil
}

// === AutoRouter Compatible Methods ===
// These methods are designed to work with the autorouter package

// Create handles trainer creation (autorouter compatible)
func (h *TrainerHandler) Create(w http.ResponseWriter, r *http.Request) {
	h.HandleCreate(w, r)
}

// Get handles trainer retrieval (autorouter compatible)
func (h *TrainerHandler) Get(w http.ResponseWriter, r *http.Request) {
	h.HandleGet(w, r)
}

// Move handles trainer movement (autorouter compatible)
func (h *TrainerHandler) Move(w http.ResponseWriter, r *http.Request) {
	h.HandleMove(w, r)
}

// FetchPosition handles position fetching (autorouter compatible)
func (h *TrainerHandler) FetchPosition(w http.ResponseWriter, r *http.Request) {
	h.HandleFetchPosition(w, r)
}

// List handles trainer listing (autorouter compatible)
func (h *TrainerHandler) List(w http.ResponseWriter, r *http.Request) {
	h.HandleList(w, r)
}

// Status handles trainer status (autorouter compatible)
func (h *TrainerHandler) Status(w http.ResponseWriter, r *http.Request) {
	h.HandleStatus(w, r)
}

