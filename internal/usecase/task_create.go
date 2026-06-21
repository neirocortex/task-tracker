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
	taskRepo TaskSaver
	tagRepo  TaskTagsSyncer
}

func NewCreateTaskCommand(taskRepo TaskSaver, tagRepo TaskTagsSyncer) *CreateTaskCommand {
	return &CreateTaskCommand{
		taskRepo: taskRepo,
		tagRepo:  tagRepo,
	}
}

func (c *CreateTaskCommand) Execute(ctx context.Context, task *domain.Task, tagNames []string) error {
	if task.DueDate.Before(time.Now()) {
		return ErrInvalidDueDate
	}
	task.Status = domain.StatusNew

	if err := c.taskRepo.Create(ctx, task); err != nil {
		return err
	}

	if len(tagNames) == 0 {
		return nil
	}

	syncTags, err := c.tagRepo.SyncTaskTags(ctx, task.ID, tagNames)
	if err != nil {
		return err
	}

	domainTags := make([]domain.Tag, len(syncTags))
	for i, name := range syncTags {
		tag, _ := domain.NewTag(name)
		domainTags[i] = tag
	}

	task.Tags = domainTags
	return nil
}
