package usecase

import (
	"context"
)

// cqrs for solid srp : every command has separate object
type DeleteTaskCommand struct {
	repo TaskRemover
}

func NewDeleteTaskCommand(repo TaskRemover) *DeleteTaskCommand {
	return &DeleteTaskCommand{repo: repo}
}

func (c *DeleteTaskCommand) Execute(ctx context.Context, id int64) error {
	if err := c.validate(); err != nil {
		return err
	}

	return c.repo.Delete(ctx, id)
}

func (c *DeleteTaskCommand) validate() error {
	return nil
}
