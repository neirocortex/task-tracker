package usecase

import (
	"context"
	"taskTracker/internal/domain"
)

// cqrs for solid srp : every command has separate object
type ListTasksQuery struct {
	taskRepo TaskViewer
	tagRepo  TaskTagsBulkViewer
}

func NewListTasksQuery(taskRepo TaskViewer, tagRepo TaskTagsBulkViewer) *ListTasksQuery {
	return &ListTasksQuery{taskRepo: taskRepo, tagRepo: tagRepo}
}

func (q *ListTasksQuery) Execute(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, error) {
	tasks, err := q.taskRepo.GetList(ctx, filter)
	if err != nil || len(tasks) == 0 {
		return tasks, err
	}

	taskIDs := make([]int64, len(tasks))
	for i, task := range tasks {
		taskIDs[i] = task.ID
	}

	tagsMap, err := q.tagRepo.FetchTagsForTasks(ctx, taskIDs)
	if err != nil {
		return nil, err
	}

	for i := range tasks {
		if tags, exists := tagsMap[tasks[i].ID]; exists {
			tasks[i].Tags = tags
		} else {
			tasks[i].Tags = []domain.Tag{}
		}
	}

	return tasks, nil
}
