package usecase

import (
	"context"
	"errors"
	"taskTracker/internal/domain"
	"time"
)

var ErrInvalidDueDate = errors.New("due date must be in the future")

// cqrs for solid srp : every command has separate object
type CreateTaskCommand struct {
	repo TaskSaver
}

func NewCreateTaskCommand(repo TaskSaver) *CreateTaskCommand {
	return &CreateTaskCommand{repo: repo}
}

func (c *CreateTaskCommand) Execute(ctx context.Context, task *domain.Task) error {
	if task.DueDate.Before(time.Now()) {
		return ErrInvalidDueDate
	}
	task.Status = domain.StatusNew
	return c.repo.Create(ctx, task)
}
