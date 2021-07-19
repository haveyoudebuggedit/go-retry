package retry

import (
	"context"
)

// ContextStrategy provides a timeout based on a context in the ctx parameter. If the context is canceled the
// retry loop is aborted.
func ContextStrategy(ctx context.Context) Strategy {
	return NewStrategy(
		func() StrategyInstance {
			return &contextStrategy{
				ctx: ctx,
			}
		},
		false,
		false,
		true,
	)
}

type contextStrategy struct {
	ctx context.Context
}

func (c *contextStrategy) Continue(wrap WrapFunc, _ error, _ string) error {
	return nil
}

func (c *contextStrategy) Wait(_ error) interface{} {
	return c.ctx.Done()
}

func (c *contextStrategy) OnWaitExpired(wrap WrapFunc, err error, action string) error {
	return wrap(
		err,
		"timeout while %s",
		action,
	)
}