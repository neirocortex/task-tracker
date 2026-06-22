package domain

import (
	"errors"
	"time"
)

type TaskStatus string

const (
	StatusNew        TaskStatus = "NEW"
	StatusInProgress TaskStatus = "IN_PROGRESS"
	StatusDone       TaskStatus = "DONE"
	StatusCanceled   TaskStatus = "CANCELED"
)

var (
	ErrTitleEmpty = errors.New("title cannot be empty")
)

// clean models, no json
type Task struct {
	ID          int64
	Title       string
	Description string
	DueDate     time.Time
	Status      TaskStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Tags        []Tag
	Recurrence  *TaskRecurrence // nil is for usual tasks
}

type TaskFilter struct {
	Status      *TaskStatus
	DueDateFrom *time.Time
	DueDateTo   *time.Time
	Limit       int
	Offset      int
}
