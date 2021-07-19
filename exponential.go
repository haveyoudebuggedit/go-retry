package retry

import (
	"time"
)

// ExponentialBackoff is a retry strategy that increases the wait time after each call by the specified factor.
func ExponentialBackoff(factor uint8) Strategy {
	return NewStrategy(
		func() StrategyInstance {
			waitTime := time.Second
			return &exponentialBackoff{
				waitTime: waitTime,
				factor:   factor,
			}
		},
		false,
		true,
		false,
	)
}

type exponentialBackoff struct {
	waitTime time.Duration
	factor   uint8
}

func (e *exponentialBackoff) Wait(_ error) interface{} {
	return time.After(e.waitTime)
}

func (e *exponentialBackoff) OnWaitExpired(_ WrapFunc, _ error, _ string) error {
	return nil
}

func (e *exponentialBackoff) Continue(_ WrapFunc, _ error, _ string) error {
	return nil
}