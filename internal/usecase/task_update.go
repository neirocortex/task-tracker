package usecase

import (
	"context"
	"taskTracker/internal/domain"
)

type UpdateTaskCommand struct {
	taskRepo TaskModifier
	tagRepo  TaskTagsSyncer
}

func NewUpdateTaskCommand(taskRepo TaskModifier, tagRepo TaskTagsSyncer) *UpdateTaskCommand {
	return &UpdateTaskCommand{taskRepo: taskRepo, tagRepo: tagRepo}
}

func (c *UpdateTaskCommand) Execute(ctx context.Context, task *domain.Task, tagNames []string) error {
	if err := c.taskRepo.Update(ctx, task); err != nil {
		return err
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
