package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"taskTracker/internal/domain"
	"taskTracker/internal/usecase"
)

type TagHandler struct {
	createCmd *usecase.CreateTagCommand
	updateCmd *usecase.UpdateTagCommand
	deleteCmd *usecase.DeleteTagCommand
	listQuery *usecase.GetTagsQuery
}

func NewTagHandler(
	createCmd *usecase.CreateTagCommand,
	updateCmd *usecase.UpdateTagCommand,
	deleteCmd *usecase.DeleteTagCommand,
	listQuery *usecase.GetTagsQuery,
) *TagHandler {
	return &TagHandler{
		createCmd: createCmd,
		updateCmd: updateCmd,
		deleteCmd: deleteCmd,
		listQuery: listQuery,
	}
}

func (h *TagHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tags", h.CreateTag)
	mux.HandleFunc("GET /api/v1/tags", h.GetTags)
	mux.HandleFunc("DELETE /api/v1/tags/{name}", h.DeleteTag)
	mux.HandleFunc("PUT /api/v1/tags/{name}", h.UpdateTag)
}

func (h *TagHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var req TagCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := req.Validate(); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.createCmd.Execute(r.Context(), req.Name); err != nil {
		if errors.Is(err, domain.ErrTagEmpty) {
			h.respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *TagHandler) GetTags(w http.ResponseWriter, r *http.Request) {
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

	offset := (page - 1) * limit

	paginatedData, err := h.listQuery.Execute(r.Context(), limit, offset)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}

	tagResponses := NewTagListResponse(paginatedData.Tags)

	totalPages := 0
	if paginatedData.TotalCount > 0 {
		totalPages = (paginatedData.TotalCount + limit - 1) / limit
	}

	response := PaginatedTagsResponse{
		Data: tagResponses,
		Pagination: TagPaginationMeta{
			CurrentPage: page,
			Limit:       limit,
			TotalItems:  paginatedData.TotalCount,
			TotalPages:  totalPages,
		},
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	tagName := r.PathValue("name")
	if tagName == "" {
		h.respondWithError(w, http.StatusBadRequest, "tag name path parameter is required")
		return
	}

	if err := h.deleteCmd.Execute(r.Context(), tagName); err != nil {
		if errors.Is(err, domain.ErrSystemTagModification) {
			h.respondWithError(w, http.StatusForbidden, err.Error())
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TagHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	oldName := r.PathValue("name")
	if oldName == "" {
		h.respondWithError(w, http.StatusBadRequest, "tag name path parameter is required")
		return
	}

	var req TagUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := req.Validate(); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.updateCmd.Execute(r.Context(), oldName, req.NewName)
	if err != nil {
		if errors.Is(err, domain.ErrSystemTagModification) {
			h.respondWithError(w, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			h.respondWithError(w, http.StatusNotFound, "tag not found")
			return
		}

		h.respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *TagHandler) respondWithError(w http.ResponseWriter, code int, msg string) {
	h.respondWithJSON(w, code, map[string]string{"error": msg})
}

func (h *TagHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
