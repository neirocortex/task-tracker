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
	TotalPages int
}

func (q *GetTagsQuery) Execute(ctx context.Context, limit int, page int) (PaginatedTags, error) {
	if err := q.validate(limit, page); err != nil {
		return PaginatedTags{}, err
	}

	tags, totalCount, err := q.repo.FindAllTags(ctx, limit, calcOffsetTag(limit, page))
	if err != nil {
		return PaginatedTags{}, err
	}
	return PaginatedTags{Tags: tags, TotalCount: totalCount, TotalPages: calcPagesTag(totalCount, limit)}, nil
}

func (q *GetTagsQuery) validate(limit, page int) error {
	if limit <= 0 || limit > 100 {
		return domain.ErrTagInvalid
	}

	if page < 0 {
		return domain.ErrTagInvalid
	}

	return nil
}

func calcOffsetTag(limit int, page int) int {
	if page > 0 {
		return (page - 1) * limit
	} else {
		return 0
	}
}

func calcPagesTag(totalCount int, limit int) int {
	if totalCount > 0 && limit > 0 {
		return (totalCount + limit - 1) / limit
	} else {
		return 0
	}
}
