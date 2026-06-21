package usecase

import (
	"context"
	"taskTracker/internal/domain"
)

type DeleteTagCommand struct {
	repo TagRemover
}

func NewDeleteTagCommand(repo TagRemover) *DeleteTagCommand {
	return &DeleteTagCommand{repo: repo}
}

func (c *DeleteTagCommand) Execute(ctx context.Context, tagName string) error {
	tag, err := domain.NewTag(tagName)
	if err != nil {
		return err
	}
	if err := tag.CanDelete(); err != nil {
		return err
	}
	return c.repo.DeleteTagFromRegistry(ctx, tag.Name)
}
