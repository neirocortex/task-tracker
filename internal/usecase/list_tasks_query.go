package usecase

import (
	"context"
	"taskTracker/internal/domain"
)

// sqrs for solid srp : every command has separate object
type ListTasksQuery struct {
	repo TaskViewer
}

func NewListTasksQuery(repo TaskViewer) *ListTasksQuery {
	return &ListTasksQuery{repo: repo}
}

func (q *ListTasksQuery) Execute(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, error) {
	return q.repo.GetList(ctx, filter)
}
