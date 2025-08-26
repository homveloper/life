package handler

import (
	"context"
	"fmt"

	"github.com/danghamo/life/internal/app/query"
	"github.com/danghamo/life/internal/domain/shared"
	"github.com/danghamo/life/internal/domain/trainer"
)

// TrainerQueryHandler handles trainer queries
type TrainerQueryHandler struct {
	trainerRepo trainer.Repository
}

// NewTrainerQueryHandler creates a new trainer query handler
func NewTrainerQueryHandler(trainerRepo trainer.Repository) *TrainerQueryHandler {
	return &TrainerQueryHandler{
		trainerRepo: trainerRepo,
	}
}

// Handle handles trainer queries
func (h *TrainerQueryHandler) Handle(ctx context.Context, q query.Query) (interface{}, error) {
	switch qu := q.(type) {
	case query.GetTrainerByIDQuery:
		return h.handleGetTrainerByID(ctx, qu)
	case query.GetTrainerByNicknameQuery:
		return h.handleGetTrainerByNickname(ctx, qu)
	case query.GetTrainersByPositionQuery:
		return h.handleGetTrainersByPosition(ctx, qu)
	case query.GetTrainerInventoryQuery:
		return h.handleGetTrainerInventory(ctx, qu)
	case query.GetTrainerInventoryByTypeQuery:
		return h.handleGetTrainerInventoryByType(ctx, qu)
	case query.GetTrainerPartyQuery:
		return h.handleGetTrainerParty(ctx, qu)
	case query.GetTrainerStatsQuery:
		return h.handleGetTrainerStats(ctx, qu)
	case query.ListTrainersQuery:
		return h.handleListTrainers(ctx, qu)
	default:
		return nil, fmt.Errorf("unknown query type: %T", q)
	}
}

func (h *TrainerQueryHandler) handleGetTrainerByID(ctx context.Context, q query.GetTrainerByIDQuery) (*trainer.Trainer, error) {
	userID := trainer.UserID(q.TrainerID)
	return h.trainerRepo.GetByID(ctx, userID)
}

func (h *TrainerQueryHandler) handleGetTrainerByNickname(ctx context.Context, q query.GetTrainerByNicknameQuery) (*trainer.Trainer, error) {
	return h.trainerRepo.FindByNickname(ctx, q.Nickname)
}

func (h *TrainerQueryHandler) handleGetTrainersByPosition(ctx context.Context, q query.GetTrainersByPositionQuery) ([]*trainer.Trainer, error) {
	return h.trainerRepo.GetByPosition(ctx, q.Position)
}

func (h *TrainerQueryHandler) handleGetTrainerInventory(ctx context.Context, q query.GetTrainerInventoryQuery) (trainer.Inventory, error) {
	userID := trainer.UserID(q.TrainerID)
	t, err := h.trainerRepo.GetByID(ctx, userID)
	if err != nil {
		return trainer.Inventory{}, err
	}
	if t == nil {
		return trainer.Inventory{}, shared.ErrNotFound("trainer")
	}

	return t.Inventory, nil
}

func (h *TrainerQueryHandler) handleGetTrainerInventoryByType(ctx context.Context, q query.GetTrainerInventoryByTypeQuery) ([]*trainer.Item, error) {
	userID := trainer.UserID(q.TrainerID)
	t, err := h.trainerRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, shared.ErrNotFound("trainer")
	}

	return t.Inventory.GetItemsByType(q.ItemType), nil
}

func (h *TrainerQueryHandler) handleGetTrainerParty(ctx context.Context, q query.GetTrainerPartyQuery) (trainer.AnimalParty, error) {
	userID := trainer.UserID(q.TrainerID)
	t, err := h.trainerRepo.GetByID(ctx, userID)
	if err != nil {
		return trainer.AnimalParty{}, err
	}
	if t == nil {
		return trainer.AnimalParty{}, shared.ErrNotFound("trainer")
	}

	return t.Party, nil
}

func (h *TrainerQueryHandler) handleGetTrainerStats(ctx context.Context, q query.GetTrainerStatsQuery) (shared.Stats, error) {
	userID := trainer.UserID(q.TrainerID)
	t, err := h.trainerRepo.GetByID(ctx, userID)
	if err != nil {
		return shared.Stats{}, err
	}
	if t == nil {
		return shared.Stats{}, shared.ErrNotFound("trainer")
	}

	return t.Stats, nil
}

type TrainerSummary struct {
	ID       trainer.UserID    `json:"id"`
	Nickname string            `json:"nickname"`
	Level    int               `json:"level"`
	Position shared.Position   `json:"position"`
}

func (h *TrainerQueryHandler) handleListTrainers(ctx context.Context, q query.ListTrainersQuery) (query.PaginatedResult, error) {
	// This is a simplified implementation
	// In a real implementation, you would use the pagination parameters
	// and potentially have a separate read model or repository method for pagination

	// For now, return empty paginated result as this would require
	// additional repository methods or read models
	return query.NewPaginatedResult([]TrainerSummary{}, q.Pagination.Page, q.Pagination.PageSize, 0), nil
}
