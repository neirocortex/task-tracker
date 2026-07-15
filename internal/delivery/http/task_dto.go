package http

import (
	"net/http"
	"strconv"
	"taskTracker/internal/domain"
	"time"
)

// DTO objects for clean domain model
type CreateTaskRequest struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	DueDate     time.Time      `json:"due_date"`
	Tags        []string       `json:"tags"`
	Recurrence  *RecurrenceDTO `json:"recurrence,omitempty"`
}

func (r *CreateTaskRequest) ToDomain() *domain.Task {
	return &domain.Task{
		Title:       r.Title,
		Description: r.Description,
		DueDate:     r.DueDate,
		Tags:        []domain.Tag{},
		Recurrence:  r.Recurrence.ToDomain(),
	}
}

type UpdateTaskRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	DueDate     time.Time         `json:"due_date"`
	Status      domain.TaskStatus `json:"status"`
	Tags        []string          `json:"tags"`
	Recurrence  *RecurrenceDTO    `json:"recurrence,omitempty"`
}

func (r *UpdateTaskRequest) ToDomain(id int64) *domain.Task {
	return &domain.Task{
		ID:          id,
		Title:       r.Title,
		Description: r.Description,
		DueDate:     r.DueDate,
		Status:      r.Status,
		Tags:        []domain.Tag{},
		Recurrence:  r.Recurrence.ToDomain(),
	}
}

type TaskResponse struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	DueDate     string            `json:"due_date"`
	Status      domain.TaskStatus `json:"status"`
	Tags        []string          `json:"tags"`
	IsRecurring bool              `json:"is_recurring"`
}

type PaginatedResponse struct {
	Data       []TaskResponse     `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

type PaginationMetadata struct {
	CurrentPage int `json:"current_page"`
	Limit       int `json:"limit"`
	TotalItems  int `json:"total_items"`
	TotalPages  int `json:"total_pages"`
}

func NewTaskResponse(t *domain.Task) TaskResponse {
	return TaskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		DueDate:     t.DueDate.Format(time.RFC3339),
		Status:      t.Status,
		Tags:        t.TagsStr(),
		IsRecurring: t.IsRecurring(),
	}
}

type GetTasksRequest struct {
	Status      string
	DueDateFrom string
	DueDateTo   string
	Page        int
	Limit       int
}

func ParseGetTasksRequest(r *http.Request) GetTasksRequest {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	return GetTasksRequest{
		Status:      q.Get("status"),
		DueDateFrom: q.Get("due_date_from"),
		DueDateTo:   q.Get("due_date_to"),
		Page:        page,
		Limit:       limit,
	}
}

func (req GetTasksRequest) ToDomainFilter() *domain.TaskFilter {
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
	filter.Limit = req.Limit
	if req.Page > 0 {
		filter.Offset = (req.Page - 1) * req.Limit
	} else {
		filter.Offset = 0
	}

	return &filter
}
