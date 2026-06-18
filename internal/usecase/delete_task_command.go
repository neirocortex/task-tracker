package usecase

import (
	"context"
)

// sqrs for solid srp : every command has separate object
type DeleteTaskCommand struct {
	repo TaskRemover
}

func NewDeleteTaskCommand(repo TaskRemover) *DeleteTaskCommand {
	return &DeleteTaskCommand{repo: repo}
}

func (c *DeleteTaskCommand) Execute(ctx context.Context, id int64) error {
	return c.repo.Delete(ctx, id)
}
