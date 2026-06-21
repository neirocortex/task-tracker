package usecase

import (
	"context"
	"taskTracker/internal/domain"
)

// solid isp/dip contracts with db
type TaskSaver interface {
	Create(ctx context.Context, task *domain.Task) error
}

type TaskModifier interface {
	Update(ctx context.Context, task *domain.Task) error
}

type TaskRemover interface {
	Delete(ctx context.Context, id int64) error
}

type TaskViewer interface {
	GetByID(ctx context.Context, id int64) (*domain.Task, error)
	GetList(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, int, error)
}
