package usecase

import (
	"context"
	"taskTracker/internal/domain"
)

// cqrs for solid srp : every command has separate object
type GetTaskByIDQuery struct {
	taskRepo TaskViewer
	tagRepo  TaskTagsBulkViewer
}

func NewGetTaskByIDQuery(taskRepo TaskViewer, tagRepo TaskTagsBulkViewer) *GetTaskByIDQuery {
	return &GetTaskByIDQuery{
		taskRepo: taskRepo,
		tagRepo:  tagRepo,
	}
}

func (q *GetTaskByIDQuery) Execute(ctx context.Context, id int64) (*domain.Task, error) {
	if err := q.validate(); err != nil {
		return nil, err
	}

	task, err := q.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tagsMap, err := q.tagRepo.FetchTagsForTasks(ctx, []int64{id})
	if err != nil {
		return nil, err
	}

	if tags, exists := tagsMap[id]; exists {
		task.Tags = tags
	} else {
		task.Tags = []domain.Tag{}
	}

	return task, nil
}

func (q *GetTaskByIDQuery) validate() error {
	return nil
}
