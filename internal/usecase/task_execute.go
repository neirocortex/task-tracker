package usecase

import (
	"context"
	"taskTracker/internal/domain"
	"time"
)

type RecordExecutionCommand struct {
	taskRepo TaskExecutionSaver
}

func NewRecordExecutionCommand(taskRepo TaskExecutionSaver) *RecordExecutionCommand {
	return &RecordExecutionCommand{taskRepo: taskRepo}
}

func (c *RecordExecutionCommand) Execute(ctx context.Context, taskID int64, date time.Time, status domain.TaskStatus) error {
	if err := c.validate(date, status); err != nil {
		return err
	}

	return c.taskRepo.SaveExecutionStatus(ctx, taskID, date, status)
}

func (c *RecordExecutionCommand) validate(date time.Time, status domain.TaskStatus) error {
	if date.IsZero() {
		return domain.ErrTaskInvalid
	}

	if _, ok := domain.StatusTypes[status]; !ok {
		return domain.ErrTaskInvalid
	}

	return nil
}
