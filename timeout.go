package retry

import (
	"time"
)

// CallTimeout is a strategy that will timeout individual API call retries.
func CallTimeout(timeout time.Duration) Strategy {
	return NewStrategy(
		func() StrategyInstance {
			startTime := time.Now()
			return &timeoutStrategy{
				duration:  timeout,
				startTime: startTime,
			}
		},
		false,
		false,
		true,
	)
}

type timeoutStrategy struct {
	duration  time.Duration
	startTime time.Time
}

func (t *timeoutStrategy) Continue(wrap WrapFunc, err error, action string) error {
	if t.startTime.Add(t.duration).Before(time.Now()) {
		return wrap(
			err,
			"timeout while %s, giving up",
			action,
		)
	}
	return nil
}

func (t *timeoutStrategy) Wait(_ error) interface{} {
	return nil
}

func (t *timeoutStrategy) OnWaitExpired(_ WrapFunc, _ error, _ string) error {
	return nil
}