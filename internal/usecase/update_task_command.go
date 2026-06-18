package usecase

import (
	"context"
	"taskTracker/internal/domain"
)

// cqrs for solid srp : every command has separate object
type UpdateTaskCommand struct {
	repo TaskModifier
}

func NewUpdateTaskCommand(repo TaskModifier) *UpdateTaskCommand {
	return &UpdateTaskCommand{repo: repo}
}

func (c *UpdateTaskCommand) Execute(ctx context.Context, task *domain.Task) error {
	return c.repo.Update(ctx, task)
}
