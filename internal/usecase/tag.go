package usecase

import (
	"context"
	"taskTracker/internal/domain"
)

type TagLister interface {
	FindAllTags(ctx context.Context) ([]domain.Tag, error)
}

type TaskTagsBulkViewer interface {
	FetchTagsForTasks(ctx context.Context, taskIDs []int64) (map[int64][]domain.Tag, error)
}

type TaskTagsSyncer interface {
	SyncTaskTags(ctx context.Context, taskID int64, tagNames []string) ([]string, error)
}

type TagSaver interface {
	CreateTagInRegistry(ctx context.Context, tag domain.Tag) error
}

type TagRemover interface {
	DeleteTagFromRegistry(ctx context.Context, tagName string) error
}
