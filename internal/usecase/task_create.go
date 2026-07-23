package usecase

import (
	"context"
	"taskTracker/internal/domain"
	"time"
)

// cqrs for solid srp : every command has separate object
type CreateTaskCommand struct {
	taskRepo         TaskSaver
	tagRepo          TaskTagsSyncer
	taskSaveNotyfier TaskSaveNotyfier
}

func NewCreateTaskCommand(taskRepo TaskSaver, tagRepo TaskTagsSyncer, taskSaveNotyfier TaskSaveNotyfier) *CreateTaskCommand {
	return &CreateTaskCommand{
		taskRepo:         taskRepo,
		tagRepo:          tagRepo,
		taskSaveNotyfier: taskSaveNotyfier,
	}
}

func (c *CreateTaskCommand) Execute(ctx context.Context, task *domain.Task, tagNames []string) error {
	if err := c.validate(task, tagNames); err != nil {
		return err
	}

	task.Status = domain.StatusNew

	if err := c.taskRepo.Create(ctx, task); err != nil {
		return err
	}

	if c.taskSaveNotyfier != nil {
		c.taskSaveNotyfier.SendCreate(ctx, task)
	}

	if len(tagNames) != 0 {
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
	}

	return nil
}

func (c *CreateTaskCommand) validate(task *domain.Task, tagNames []string) error {
	if task.Title == "" {
		return domain.ErrTaskInvalid
	}

	if task.DueDate.Before(time.Now()) {
		return domain.ErrTaskInvalid
	}

	if task.Recurrence != nil {
		if _, ok := domain.ReccurenceTypes[task.Recurrence.Type]; !ok {
			return domain.ErrTaskInvalid
		}
	}

	for _, tagName := range tagNames {
		if tagName == "" {
			return domain.ErrTaskInvalid
		}
	}

	return nil
}
