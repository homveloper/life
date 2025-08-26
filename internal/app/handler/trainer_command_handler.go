package handler

import (
	"context"
	"fmt"

	"github.com/danghamo/life/internal/app/command"
	"github.com/danghamo/life/internal/domain/shared"
	"github.com/danghamo/life/internal/domain/trainer"
)

// TrainerCommandHandler handles trainer commands
type TrainerCommandHandler struct {
	trainerRepo trainer.Repository
}

// NewTrainerCommandHandler creates a new trainer command handler
func NewTrainerCommandHandler(trainerRepo trainer.Repository) *TrainerCommandHandler {
	return &TrainerCommandHandler{
		trainerRepo: trainerRepo,
	}
}

// Handle handles trainer commands
func (h *TrainerCommandHandler) Handle(ctx context.Context, cmd command.Command) error {
	switch c := cmd.(type) {
	case command.CreateTrainerCommand:
		return h.handleCreateTrainer(ctx, c)
	case command.MoveTrainerCommand:
		return h.handleMoveTrainer(ctx, c)
	case command.AddAnimalToPartyCommand:
		return h.handleAddAnimalToParty(ctx, c)
	case command.RemoveAnimalFromPartyCommand:
		return h.handleRemoveAnimalFromParty(ctx, c)
	case command.AddItemToInventoryCommand:
		return h.handleAddItemToInventory(ctx, c)
	case command.RemoveItemFromInventoryCommand:
		return h.handleRemoveItemFromInventory(ctx, c)
	case command.GainExperienceCommand:
		return h.handleGainExperience(ctx, c)
	case command.SpendMoneyCommand:
		return h.handleSpendMoney(ctx, c)
	case command.EarnMoneyCommand:
		return h.handleEarnMoney(ctx, c)
	default:
		return fmt.Errorf("unknown command type: %T", cmd)
	}
}

func (h *TrainerCommandHandler) handleCreateTrainer(ctx context.Context, cmd command.CreateTrainerCommand) error {
	nickname, err := trainer.NewNickname(cmd.Nickname)
	if err != nil {
		return err
	}

	userID := trainer.UserID(cmd.CommandID())
	return h.trainerRepo.FindOneAndInsert(ctx, userID, func() (*trainer.Trainer, error) {
		return trainer.NewTrainer(userID, nickname)
	})
}

func (h *TrainerCommandHandler) handleMoveTrainer(ctx context.Context, cmd command.MoveTrainerCommand) error {
	userID := trainer.UserID(cmd.TrainerID)

	return h.trainerRepo.FindOneAndUpdate(ctx, userID, func(t *trainer.Trainer) (*trainer.Trainer, error) {
		if t == nil {
			return nil, shared.ErrNotFound("trainer")
		}

		err := t.MoveTo(cmd.NewPosition)
		if err != nil {
			return nil, err
		}

		return t, nil
	})
}

func (h *TrainerCommandHandler) handleAddAnimalToParty(ctx context.Context, cmd command.AddAnimalToPartyCommand) error {
	userID := trainer.UserID(cmd.TrainerID)
	animalID := shared.ID(cmd.AnimalID)

	return h.trainerRepo.FindOneAndUpdate(ctx, userID, func(t *trainer.Trainer) (*trainer.Trainer, error) {
		if t == nil {
			return nil, shared.ErrNotFound("trainer")
		}

		err := t.AddAnimalToParty(animalID)
		if err != nil {
			return nil, err
		}

		return t, nil
	})
}

func (h *TrainerCommandHandler) handleRemoveAnimalFromParty(ctx context.Context, cmd command.RemoveAnimalFromPartyCommand) error {
	userID := trainer.UserID(cmd.TrainerID)
	animalID := shared.ID(cmd.AnimalID)

	return h.trainerRepo.FindOneAndUpdate(ctx, userID, func(t *trainer.Trainer) (*trainer.Trainer, error) {
		if t == nil {
			return nil, shared.ErrNotFound("trainer")
		}

		err := t.RemoveAnimalFromParty(animalID)
		if err != nil {
			return nil, err
		}

		return t, nil
	})
}

func (h *TrainerCommandHandler) handleAddItemToInventory(ctx context.Context, cmd command.AddItemToInventoryCommand) error {
	userID := trainer.UserID(cmd.TrainerID)

	return h.trainerRepo.FindOneAndUpdate(ctx, userID, func(t *trainer.Trainer) (*trainer.Trainer, error) {
		if t == nil {
			return nil, shared.ErrNotFound("trainer")
		}

		item, err := trainer.NewItem(cmd.ItemType, cmd.ItemName)
		if err != nil {
			return nil, err
		}

		err = t.Inventory.AddItem(item)
		if err != nil {
			return nil, err
		}

		return t, nil
	})
}

func (h *TrainerCommandHandler) handleRemoveItemFromInventory(ctx context.Context, cmd command.RemoveItemFromInventoryCommand) error {
	userID := trainer.UserID(cmd.TrainerID)
	itemID := trainer.ItemID(cmd.ItemID)

	return h.trainerRepo.FindOneAndUpdate(ctx, userID, func(t *trainer.Trainer) (*trainer.Trainer, error) {
		if t == nil {
			return nil, shared.ErrNotFound("trainer")
		}

		_, err := t.Inventory.RemoveItem(itemID)
		if err != nil {
			return nil, err
		}

		return t, nil
	})
}

func (h *TrainerCommandHandler) handleGainExperience(ctx context.Context, cmd command.GainExperienceCommand) error {
	userID := trainer.UserID(cmd.TrainerID)

	return h.trainerRepo.FindOneAndUpdate(ctx, userID, func(t *trainer.Trainer) (*trainer.Trainer, error) {
		if t == nil {
			return nil, shared.ErrNotFound("trainer")
		}

		err := t.GainExperience(cmd.Points)
		if err != nil {
			return nil, err
		}

		return t, nil
	})
}

func (h *TrainerCommandHandler) handleSpendMoney(ctx context.Context, cmd command.SpendMoneyCommand) error {
	userID := trainer.UserID(cmd.TrainerID)

	return h.trainerRepo.FindOneAndUpdate(ctx, userID, func(t *trainer.Trainer) (*trainer.Trainer, error) {
		if t == nil {
			return nil, shared.ErrNotFound("trainer")
		}

		err := t.SpendMoney(cmd.Amount)
		if err != nil {
			return nil, err
		}

		return t, nil
	})
}

func (h *TrainerCommandHandler) handleEarnMoney(ctx context.Context, cmd command.EarnMoneyCommand) error {
	userID := trainer.UserID(cmd.TrainerID)

	return h.trainerRepo.FindOneAndUpdate(ctx, userID, func(t *trainer.Trainer) (*trainer.Trainer, error) {
		if t == nil {
			return nil, shared.ErrNotFound("trainer")
		}

		err := t.EarnMoney(cmd.Amount)
		if err != nil {
			return nil, err
		}

		return t, nil
	})
}
