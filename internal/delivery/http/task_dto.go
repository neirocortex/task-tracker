package http

import (
	"net/http"
	"taskTracker/internal/domain"
	"time"
)

// DTO objects for clean domain model
type CreateTaskRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Tags        []string  `json:"tags"`
}

func (r *CreateTaskRequest) Validate() error {
	if r.Title == "" {
		return stringError("title is required")
	}
	return nil
}

func (r *CreateTaskRequest) ToDomain() *domain.Task {
	return &domain.Task{
		Title:       r.Title,
		Description: r.Description,
		DueDate:     r.DueDate,
		Tags:        []domain.Tag{},
	}
}

type UpdateTaskRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	DueDate     time.Time         `json:"due_date"`
	Status      domain.TaskStatus `json:"status"`
	Tags        []string          `json:"tags"`
}

func (r *UpdateTaskRequest) ToDomain(id int64) *domain.Task {
	return &domain.Task{
		ID:          id,
		Title:       r.Title,
		Description: r.Description,
		DueDate:     r.DueDate,
		Status:      r.Status,
		Tags:        []domain.Tag{},
	}
}

type TaskResponse struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	DueDate     string            `json:"due_date"`
	Status      domain.TaskStatus `json:"status"`
	Tags        []TagResponse     `json:"tags"`
}

func NewTaskResponse(t *domain.Task) TaskResponse {
	dtoTags := make([]TagResponse, len(t.Tags))
	for i, tag := range t.Tags {
		dtoTags[i] = TagResponse{
			Name:     tag.Name,
			IsSystem: tag.IsSystem,
		}
	}

	return TaskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		DueDate:     t.DueDate.Format(time.RFC3339),
		Status:      t.Status,
		Tags:        dtoTags,
	}
}

type GetTasksRequest struct {
	Status      string
	DueDateFrom string
	DueDateTo   string
}

func ParseGetTasksRequest(r *http.Request) GetTasksRequest {
	q := r.URL.Query()
	return GetTasksRequest{
		Status:      q.Get("status"),
		DueDateFrom: q.Get("due_date_from"),
		DueDateTo:   q.Get("due_date_to"),
	}
}

func (req GetTasksRequest) ToDomainFilter() domain.TaskFilter {
	filter := domain.TaskFilter{}
	if req.Status != "" {
		s := domain.TaskStatus(req.Status)
		filter.Status = &s
	}
	if t, err := time.Parse(time.RFC3339, req.DueDateFrom); err == nil {
		filter.DueDateFrom = &t
	}
	if t, err := time.Parse(time.RFC3339, req.DueDateTo); err == nil {
		filter.DueDateTo = &t
	}
	return filter
}

type stringError string

func (e stringError) Error() string { return string(e) }
