package usecase

import (
	"context"
	"taskTracker/internal/domain"
)

type GetTagsQuery struct {
	repo TagLister
}

func NewGetTagsQuery(repo TagLister) *GetTagsQuery {
	return &GetTagsQuery{repo: repo}
}

func (q *GetTagsQuery) Execute(ctx context.Context) ([]domain.Tag, error) {
	return q.repo.FindAllTags(ctx)
}
