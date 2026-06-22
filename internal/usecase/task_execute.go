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
	return c.taskRepo.SaveExecutionStatus(ctx, taskID, date, status)
}
