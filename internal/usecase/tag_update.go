package usecase

import (
	"context"
	"taskTracker/internal/domain"
)

type UpdateTagCommand struct {
	repo TagModifier
}

func NewUpdateTagCommand(repo TagModifier) *UpdateTagCommand {
	return &UpdateTagCommand{repo: repo}
}

func (c *UpdateTagCommand) Execute(ctx context.Context, oldName string, newName string) error {
	if err := c.validate(oldName, newName); err != nil {
		return err
	}

	oldTag, err := domain.NewTag(oldName)
	if err != nil {
		return err
	}
	if err := oldTag.CanDelete(); err != nil {
		return err
	}

	return c.repo.UpdateTagInRegistry(ctx, oldName, newName)
}

func (c *UpdateTagCommand) validate(oldName string, newName string) error {
	if oldName == "" || newName == "" {
		return domain.ErrTagInvalid
	}

	return nil
}
