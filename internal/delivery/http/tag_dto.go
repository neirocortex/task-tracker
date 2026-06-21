package http

import (
	"taskTracker/internal/domain"
)

type TagCreateRequest struct {
	Name string `json:"name"`
}

func (r *TagCreateRequest) Validate() error {
	if r.Name == "" {
		return domain.ErrTagEmpty
	}
	return nil
}

type TagResponse struct {
	Name     string `json:"name"`
	IsSystem bool   `json:"is_system"`
}

func NewTagResponse(t domain.Tag) TagResponse {
	return TagResponse{
		Name:     t.Name,
		IsSystem: t.IsSystem,
	}
}

func NewTagListResponse(tags []domain.Tag) []TagResponse {
	response := make([]TagResponse, len(tags))
	for i, tag := range tags {
		response[i] = NewTagResponse(tag)
	}
	return response
}

type TagUpdateRequest struct {
	NewName string `json:"new_name"`
}

func (r *TagUpdateRequest) Validate() error {
	if r.NewName == "" {
		return domain.ErrTagEmpty
	}
	return nil
}
