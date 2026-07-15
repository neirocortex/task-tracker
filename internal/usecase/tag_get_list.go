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

type PaginatedTags struct {
	Tags       []domain.Tag
	TotalCount int
}

func (q *GetTagsQuery) Execute(ctx context.Context, limit, offset int) (PaginatedTags, error) {
	if err := q.validate(limit, offset); err != nil {
		return PaginatedTags{}, err
	}

	tags, totalCount, err := q.repo.FindAllTags(ctx, limit, offset)
	if err != nil {
		return PaginatedTags{}, err
	}
	return PaginatedTags{Tags: tags, TotalCount: totalCount}, nil
}

func (q *GetTagsQuery) validate(limit, offset int) error {
	if limit <= 0 || limit > 100 {
		return domain.ErrTagInvalid
	}

	if offset < 0 {
		return domain.ErrTagInvalid
	}

	return nil
}
