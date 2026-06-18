package usecase

import (
	"context"
	"taskTracker/internal/domain"
)

// cqrs for solid srp : every command has separate object
type GetTaskByIDQuery struct {
	repo TaskViewer
}

func NewGetTaskByIDQuery(repo TaskViewer) *GetTaskByIDQuery {
	return &GetTaskByIDQuery{repo: repo}
}

func (q *GetTaskByIDQuery) Execute(ctx context.Context, id int64) (*domain.Task, error) {
	return q.repo.GetByID(ctx, id)
}
