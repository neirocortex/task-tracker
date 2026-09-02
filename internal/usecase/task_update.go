package usecase

import (
	"context"
	"taskTracker/internal/domain"
)

type UpdateTaskCommand struct {
	taskRepo TaskModifier
	tagRepo  TaskTagsSyncer
	cache    TaskCacheRepository
}

func NewUpdateTaskCommand(taskRepo TaskModifier, tagRepo TaskTagsSyncer, cache TaskCacheRepository) *UpdateTaskCommand {
	return &UpdateTaskCommand{taskRepo: taskRepo, tagRepo: tagRepo, cache: cache}
}

func (c *UpdateTaskCommand) Execute(ctx context.Context, task *domain.Task, tagNames []string) error {
	if c.cache != nil {
		_ = c.cache.InvalidateCalendar(ctx)
	}
	return c.executeImpl(ctx, task, tagNames)
}

func (c *UpdateTaskCommand) executeImpl(ctx context.Context, task *domain.Task, tagNames []string) error {
	if err := (&CreateTaskCommand{}).validate(task, tagNames); err != nil {
		return err
	}

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
