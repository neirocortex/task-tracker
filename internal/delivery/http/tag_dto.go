package http

import (
	"taskTracker/internal/domain"
)

type TagCreateRequest struct {
	Name string `json:"name"`
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

type PaginatedTagsResponse struct {
	Data       []TagResponse     `json:"data"`
	Pagination TagPaginationMeta `json:"pagination"`
}

type TagPaginationMeta struct {
	CurrentPage int `json:"current_page"`
	Limit       int `json:"limit"`
	TotalItems  int `json:"total_items"`
	TotalPages  int `json:"total_pages"`
}
