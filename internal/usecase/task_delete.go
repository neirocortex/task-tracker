package usecase

import (
	"context"
)

// cqrs for solid srp : every command has separate object
type DeleteTaskCommand struct {
	repo  TaskRemover
	cache TaskCacheRepository
}

func NewDeleteTaskCommand(repo TaskRemover, cache TaskCacheRepository) *DeleteTaskCommand {
	return &DeleteTaskCommand{repo: repo, cache: cache}
}

func (c *DeleteTaskCommand) Execute(ctx context.Context, id int64) error {
	if c.cache != nil {
		_ = c.cache.InvalidateCalendar(ctx)
	}
	return c.executeImpl(ctx, id)
}

func (c *DeleteTaskCommand) executeImpl(ctx context.Context, id int64) error {
	if err := c.validate(); err != nil {
		return err
	}

	return c.repo.Delete(ctx, id)
}

func (c *DeleteTaskCommand) validate() error {
	return nil
}
