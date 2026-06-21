package http

import (
	"encoding/json"
	"net/http"
	"taskTracker/internal/domain"
	"taskTracker/internal/usecase"
)

type TagHandler struct {
	createCmd *usecase.CreateTagCommand
	deleteCmd *usecase.DeleteTagCommand
	listQuery *usecase.GetTagsQuery
}

func NewTagHandler(
	createCmd *usecase.CreateTagCommand,
	deleteCmd *usecase.DeleteTagCommand,
	listQuery *usecase.GetTagsQuery,
) *TagHandler {
	return &TagHandler{
		createCmd: createCmd,
		deleteCmd: deleteCmd,
		listQuery: listQuery,
	}
}

func (h *TagHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/tags", h.CreateTag)
	mux.HandleFunc("GET /api/v1/tags", h.GetTags)
	mux.HandleFunc("DELETE /api/v1/tags/{name}", h.DeleteTag)
}

func (h *TagHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var req TagCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.createCmd.Execute(r.Context(), req.Name); err != nil {
		if err == domain.ErrTagEmpty {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *TagHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	tags, err := h.listQuery.Execute(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := NewTagListResponse(tags)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	tagName := r.PathValue("name")
	if tagName == "" {
		http.Error(w, "tag name path parameter is required", http.StatusBadRequest)
		return
	}

	if err := h.deleteCmd.Execute(r.Context(), tagName); err != nil {
		if err == domain.ErrSystemTagModification {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
