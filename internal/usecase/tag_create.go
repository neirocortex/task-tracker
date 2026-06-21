package usecase

import (
	"context"
	"taskTracker/internal/domain"
)

type CreateTagCommand struct {
	repo TagSaver
}

func NewCreateTagCommand(repo TagSaver) *CreateTagCommand {
	return &CreateTagCommand{repo: repo}
}

func (c *CreateTagCommand) Execute(ctx context.Context, tagName string) error {
	tag, err := domain.NewTag(tagName)
	if err != nil {
		return err
	}

	return c.repo.CreateTagInRegistry(ctx, tag)
}
