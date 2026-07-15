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
	if err := c.validate(tagName); err != nil {
		return err
	}

	tag, err := domain.NewTag(tagName)
	if err != nil {
		return err
	}

	return c.repo.CreateTagInRegistry(ctx, tag)
}

func (c *CreateTagCommand) validate(tagName string) error {
	if tagName == "" {
		return domain.ErrTagInvalid
	}

	return nil
}
