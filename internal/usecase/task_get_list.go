package usecase

import (
	"context"
	"sort"
	"taskTracker/internal/domain"
	"time"
)

// cqrs for solid srp : every command has separate object
type ListTasksQuery struct {
	taskRepo      TaskViewer
	tagRepo       TaskTagsBulkViewer
	executionRepo TaskExecutionViewer
}

func NewListTasksQuery(taskRepo TaskViewer, tagRepo TaskTagsBulkViewer, execRepo TaskExecutionViewer) *ListTasksQuery {
	return &ListTasksQuery{
		taskRepo:      taskRepo,
		tagRepo:       tagRepo,
		executionRepo: execRepo,
	}
}

type PaginatedTasks struct {
	Tasks      []domain.Task
	TotalCount int
	TotalPages int
}

func (q *ListTasksQuery) Execute(ctx context.Context, filter *domain.TaskFilter, limit int, page int) (PaginatedTasks, error) {
	if err := q.validate(filter, limit, page); err != nil {
		return PaginatedTasks{}, err
	}

	baseTasks, err := q.taskRepo.GetList(ctx, filter)
	if err != nil {
		return PaginatedTasks{}, err
	}
	if len(baseTasks) == 0 {
		return PaginatedTasks{Tasks: []domain.Task{}, TotalCount: 0, TotalPages: 0}, nil
	}
	taskIDs := make([]int64, len(baseTasks))
	for i, t := range baseTasks {
		taskIDs[i] = t.ID
	}
	var from, to time.Time
	var executionsMap map[int64]map[string]domain.TaskStatus
	if filter.DueDateFrom != nil && filter.DueDateTo != nil {
		from = *filter.DueDateFrom
		to = *filter.DueDateTo

		executionsMap, err = q.executionRepo.FetchExecutionsForPeriod(ctx, taskIDs, from, to)
		if err != nil {
			return PaginatedTasks{}, err
		}
	} else {
		executionsMap = make(map[int64]map[string]domain.TaskStatus)
	}
	tagsMap, err := q.tagRepo.FetchTagsForTasks(ctx, taskIDs)
	if err != nil {
		return PaginatedTasks{}, err
	}
	var virtualTasks []domain.Task
	if filter.DueDateFrom == nil || filter.DueDateTo == nil {
		virtualTasks = baseTasks
	} else {
		for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
			currentDay := d
			for _, task := range baseTasks {
				if !task.IsRecurring() {
					if task.DueDate.Year() == currentDay.Year() && task.DueDate.Month() == currentDay.Month() && task.DueDate.Day() == currentDay.Day() {
						virtualTasks = append(virtualTasks, task)
					}
					continue
				}
				if task.Recurrence.IsMatch(task.DueDate, currentDay) {
					vTask := task
					vTask.DueDate = time.Date(currentDay.Year(), currentDay.Month(), currentDay.Day(), task.DueDate.Hour(), task.DueDate.Minute(), task.DueDate.Second(), task.DueDate.Nanosecond(), task.DueDate.Location())
					dateKey := currentDay.Format(time.DateOnly)
					if taskExecs, exists := executionsMap[task.ID]; exists {
						if specificStatus, hasStatus := taskExecs[dateKey]; hasStatus {
							vTask.Status = specificStatus
						} else {
							vTask.Status = domain.StatusNew
						}
					} else {
						vTask.Status = domain.StatusNew
					}
					virtualTasks = append(virtualTasks, vTask)
				}
			}
		}
	}
	sort.SliceStable(virtualTasks, func(i, j int) bool {
		if virtualTasks[i].DueDate.Equal(virtualTasks[j].DueDate) {
			return virtualTasks[i].ID < virtualTasks[j].ID
		}
		return virtualTasks[i].DueDate.Before(virtualTasks[j].DueDate)
	})
	totalCount := len(virtualTasks)

	offset := calcOffsetTask(limit, page)

	if offset >= totalCount {
		return PaginatedTasks{Tasks: []domain.Task{}, TotalCount: totalCount, TotalPages: calcPagesTask(totalCount, limit)}, nil
	}

	end := offset + limit
	if end > totalCount {
		end = totalCount
	}

	paginatedVirtualTasks := virtualTasks[offset:end]
	for i := range paginatedVirtualTasks {
		if tags, exists := tagsMap[paginatedVirtualTasks[i].ID]; exists {
			paginatedVirtualTasks[i].Tags = tags
		} else {
			paginatedVirtualTasks[i].Tags = []domain.Tag{}
		}
	}
	return PaginatedTasks{Tasks: paginatedVirtualTasks, TotalCount: totalCount, TotalPages: calcPagesTask(totalCount, limit)}, nil
}

func (q *ListTasksQuery) validate(filter *domain.TaskFilter, limit int, page int) error {
	if limit <= 0 || limit > 100 {
		return domain.ErrTaskInvalid
	}

	if page < 0 {
		return domain.ErrTaskInvalid
	}

	if filter.DueDateFrom != nil && filter.DueDateTo != nil {
		if filter.DueDateTo.IsZero() || filter.DueDateFrom.IsZero() || filter.DueDateTo.Before(*filter.DueDateFrom) {
			return domain.ErrTaskInvalid
		}
	}

	if filter.Status != nil {
		if _, ok := domain.StatusTypes[*filter.Status]; !ok {
			return domain.ErrTaskInvalid
		}
	}

	return nil
}

func calcOffsetTask(limit, page int) int {
	if page > 0 {
		return (page - 1) * limit
	} else {
		return 0
	}
}

func calcPagesTask(totalCount int, limit int) int {
	if totalCount > 0 && limit > 0 {
		return (totalCount + limit - 1) / limit
	} else {
		return 0
	}
}
