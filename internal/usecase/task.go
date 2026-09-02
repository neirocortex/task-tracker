package usecase

import (
	"context"
	"taskTracker/internal/domain"
	"time"
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
	GetList(ctx context.Context, filter *domain.TaskFilter) ([]domain.Task, error)
}

type TaskExecutionViewer interface {
	FetchExecutionsForPeriod(ctx context.Context, taskIDs []int64, from, to time.Time) (map[int64]map[string]domain.TaskStatus, error)
}

type TaskExecutionSaver interface {
	SaveExecutionStatus(ctx context.Context, taskID int64, date time.Time, status domain.TaskStatus) error
}

// notifyer contracts
type TaskSaveNotyfier interface {
	SendCreate(ctx context.Context, task *domain.Task)
}

// cashing contracts
type ListTasksQueryExecutor interface {
	Execute(ctx context.Context, filter *domain.TaskFilter, limit int, page int) (PaginatedTasks, error)
}

type TaskCacheRepository interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	InvalidateCalendar(ctx context.Context) error
}
