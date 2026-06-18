package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"taskTracker/internal/repository/postgres"
	"taskTracker/internal/usecase"
)

// decomposition of actions for SOLID SRP
type TaskHandler struct {
	createCmd  *usecase.CreateTaskCommand
	updateCmd  *usecase.UpdateTaskCommand
	deleteCmd  *usecase.DeleteTaskCommand
	getTaskQ   *usecase.GetTaskByIDQuery
	listTasksQ *usecase.ListTasksQuery
}

func NewTaskHandler(
	createCmd *usecase.CreateTaskCommand,
	updateCmd *usecase.UpdateTaskCommand,
	deleteCmd *usecase.DeleteTaskCommand,
	getTaskQ *usecase.GetTaskByIDQuery,
	listTasksQ *usecase.ListTasksQuery,
) *TaskHandler {
	return &TaskHandler{
		createCmd:  createCmd,
		updateCmd:  updateCmd,
		deleteCmd:  deleteCmd,
		getTaskQ:   getTaskQ,
		listTasksQ: listTasksQ,
	}
}

// RESTful API: resourse oriented
func (h *TaskHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tasks", h.CreateTask)
	mux.HandleFunc("GET /api/v1/tasks", h.GetTasks)
	mux.HandleFunc("GET /api/v1/tasks/{id}", h.GetTaskByID)
	mux.HandleFunc("PUT /api/v1/tasks/{id}", h.UpdateTask)
	mux.HandleFunc("DELETE /api/v1/tasks/{id}", h.DeleteTask)
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := req.Validate(); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	task := req.ToDomain()
	if err := h.createCmd.Execute(r.Context(), task); err != nil {
		if errors.Is(err, usecase.ErrInvalidDueDate) {
			h.respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	h.respondWithJSON(w, http.StatusCreated, NewTaskResponse(task))
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	req := ParseGetTasksRequest(r)

	tasks, err := h.listTasksQ.Execute(r.Context(), req.ToDomainFilter())
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}

	response := make([]TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		response = append(response, NewTaskResponse(&t))
	}
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid id")
		return
	}

	task, err := h.getTaskQ.Execute(r.Context(), id)
	if errors.Is(err, postgres.ErrTaskNotFound) {
		h.respondWithError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	h.respondWithJSON(w, http.StatusOK, NewTaskResponse(task))
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := h.updateCmd.Execute(r.Context(), req.ToDomain(id)); err != nil {
		if errors.Is(err, postgres.ErrTaskNotFound) {
			h.respondWithError(w, http.StatusNotFound, "task not found")
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	h.respondWithJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.deleteCmd.Execute(r.Context(), id); err != nil {
		if errors.Is(err, postgres.ErrTaskNotFound) {
			h.respondWithError(w, http.StatusNotFound, "task not found")
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) respondWithError(w http.ResponseWriter, code int, msg string) {
	h.respondWithJSON(w, code, map[string]string{"error": msg})
}

func (h *TaskHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
