package grpc

import (
	"context"
	"taskTracker/internal/usecase"
	"time"

	taskv1 "taskTracker/internal/delivery/grpc/v1"
	"taskTracker/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var pbToDomainStatus = map[taskv1.TaskStatus]domain.TaskStatus{
	taskv1.TaskStatus_TASK_STATUS_NEW:         domain.StatusNew,
	taskv1.TaskStatus_TASK_STATUS_IN_PROGRESS: domain.StatusInProgress,
	taskv1.TaskStatus_TASK_STATUS_DONE:        domain.StatusDone,
	taskv1.TaskStatus_TASK_STATUS_CANCELED:    domain.StatusCanceled,
}

var domainToPbStatus = map[domain.TaskStatus]taskv1.TaskStatus{
	domain.StatusNew:        taskv1.TaskStatus_TASK_STATUS_NEW,
	domain.StatusInProgress: taskv1.TaskStatus_TASK_STATUS_IN_PROGRESS,
	domain.StatusDone:       taskv1.TaskStatus_TASK_STATUS_DONE,
	domain.StatusCanceled:   taskv1.TaskStatus_TASK_STATUS_CANCELED,
}

type TaskHandler struct {
	taskv1.UnimplementedTaskServiceServer

	createCmd     *usecase.CreateTaskCommand
	updateCmd     *usecase.UpdateTaskCommand
	deleteCmd     *usecase.DeleteTaskCommand
	getTaskQ      *usecase.GetTaskByIDQuery
	listTasksQ    *usecase.ListTasksQuery
	recordExecCmd *usecase.RecordExecutionCommand
}

func NewTaskHandler(
	createCmd *usecase.CreateTaskCommand,
	updateCmd *usecase.UpdateTaskCommand,
	deleteCmd *usecase.DeleteTaskCommand,
	getTaskQ *usecase.GetTaskByIDQuery,
	listTasksQ *usecase.ListTasksQuery,
	recordExecCmd *usecase.RecordExecutionCommand,
) *TaskHandler {
	return &TaskHandler{
		createCmd:     createCmd,
		updateCmd:     updateCmd,
		deleteCmd:     deleteCmd,
		getTaskQ:      getTaskQ,
		listTasksQ:    listTasksQ,
		recordExecCmd: recordExecCmd,
	}
}

func mapDomainErrorToGRPC(err error) error {
	switch err {
	case domain.ErrTitleEmpty, domain.ErrDateReq, domain.ErrStatusReq, domain.ErrWrongRec:
		return status.Error(codes.InvalidArgument, err.Error())
	case domain.ErrTaskNotFound:
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

func toPbTask(task *domain.Task) *taskv1.Task {
	if task == nil {
		return nil
	}

	return &taskv1.Task{
		Id:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		DueDate:     timestamppb.New(task.DueDate),
		Status:      domainToPbStatus[task.Status],
		Tags:        task.TagsStr(),
		IsRecurring: task.IsRecurring(),
	}
}

func toDomainRecurrence(taskRecurrence *taskv1.TaskRecurrence) *domain.TaskRecurrence {
	if taskRecurrence == nil {
		return nil
	}

	domainRecurrence := &domain.TaskRecurrence{
		Type:       domain.RecurrenceType(taskRecurrence.Type.String()),
		Interval:   int(taskRecurrence.Interval),
		DayOfMonth: int(taskRecurrence.DayOfMonth),
	}

	if len(taskRecurrence.Specifics) > 0 {
		domainRecurrence.Specifics = make([]time.Time, len(taskRecurrence.Specifics))
		for i, ts := range taskRecurrence.Specifics {
			domainRecurrence.Specifics[i] = ts.AsTime()
		}
	}

	return domainRecurrence
}

func toDomainTaskCreate(task *taskv1.CreateTaskRequest) *domain.Task {
	if task == nil {
		return nil
	}

	domainTask := &domain.Task{
		Title:       task.Title,
		Description: task.Description,
		DueDate:     task.DueDate.AsTime(),
		Recurrence:  toDomainRecurrence(task.Recurrence),
	}

	return domainTask
}

func toDomainTaskUpdate(task *taskv1.UpdateTaskRequest) *domain.Task {
	if task == nil {
		return nil
	}

	domainTask := &domain.Task{
		Title:       task.Title,
		Description: task.Description,
		DueDate:     task.DueDate.AsTime(),
		Recurrence:  toDomainRecurrence(task.Recurrence),
		Status:      pbToDomainStatus[task.Status],
	}

	return domainTask
}

func toDomainFilter(req *taskv1.GetTasksRequest) *domain.TaskFilter {
	filter := domain.TaskFilter{}
	if req.Status != taskv1.TaskStatus_TASK_STATUS_UNSPECIFIED {
		s := pbToDomainStatus[req.Status]
		filter.Status = &s
	}
	if req.DueDateFrom != nil {
		t := req.DueDateFrom.AsTime()
		filter.DueDateFrom = &t
	}
	if req.DueDateTo != nil {
		t := req.DueDateTo.AsTime()
		filter.DueDateTo = &t
	}

	filter.Limit = int(req.Limit)

	if req.Page > 0 {
		filter.Offset = int((req.Page - 1) * req.Limit)
	} else {
		filter.Offset = 0
	}

	return &filter
}

func (h *TaskHandler) CreateTask(ctx context.Context, req *taskv1.CreateTaskRequest) (*taskv1.CreateTaskResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	taskDomain := toDomainTaskCreate(req)

	err := h.createCmd.Execute(ctx, taskDomain, req.Tags)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &taskv1.CreateTaskResponse{
		Task: toPbTask(taskDomain),
	}, nil
}

func (h *TaskHandler) UpdateTask(ctx context.Context, req *taskv1.UpdateTaskRequest) (*taskv1.UpdateTaskResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	taskDomain := toDomainTaskUpdate(req)
	err := h.updateCmd.Execute(ctx, taskDomain, req.Tags)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &taskv1.UpdateTaskResponse{}, nil
}

func (h *TaskHandler) DeleteTask(ctx context.Context, req *taskv1.DeleteTaskRequest) (*taskv1.DeleteTaskResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	err := h.deleteCmd.Execute(ctx, req.Id)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &taskv1.DeleteTaskResponse{}, nil
}

func (h *TaskHandler) GetTaskByID(ctx context.Context, req *taskv1.GetTaskByIDRequest) (*taskv1.GetTaskByIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	taskDomain, err := h.getTaskQ.Execute(ctx, req.Id)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	return &taskv1.GetTaskByIDResponse{
		Task: toPbTask(taskDomain),
	}, nil
}

func (h *TaskHandler) GetTasks(ctx context.Context, req *taskv1.GetTasksRequest) (*taskv1.GetTasksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	filter := toDomainFilter(req)

	paginatedData, err := h.listTasksQ.Execute(ctx, filter)
	if err != nil {
		return nil, mapDomainErrorToGRPC(err)
	}

	tasksPb := make([]*taskv1.Task, len(paginatedData.Tasks))
	for i := range paginatedData.Tasks {
		tasksPb[i] = toPbTask(&paginatedData.Tasks[i])
	}

	var totalCount int64 = int64(paginatedData.TotalCount)
	var limit int64 = int64(filter.Limit)
	var page int64 = int64(filter.Offset/filter.Limit + 1)

	var totalPages int64
	if totalCount > 0 {
		totalPages = (totalCount + limit - 1) / limit
	}

	response := &taskv1.GetTasksResponse{
		Data: tasksPb,
		Pagination: &taskv1.PaginationMetadata{
			CurrentPage: page,
			Limit:       limit,
			TotalItems:  totalCount,
			TotalPages:  totalPages,
		},
	}
	return response, nil
}
