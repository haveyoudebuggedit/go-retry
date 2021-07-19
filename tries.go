package retry

// MaxTries is a strategy that will timeout individual API calls based on a maximum number of retries. The total number
// of API calls can be higher in case of a complex functions that involve multiple API calls.
func MaxTries(tries uint16) Strategy {
	return NewStrategy(
		func() StrategyInstance {
			return &maxTriesStrategy{
				maxTries: tries,
				tries:    0,
			}
		},
		false,
		false,
		true,
	)
}

type maxTriesStrategy struct {
	maxTries uint16
	tries    uint16
}

func (m *maxTriesStrategy) Continue(wrap WrapFunc, err error, action string) error {
	m.tries++
	if m.tries > m.maxTries {
		return wrap(
			err,
			"maximum retries reached while %s, giving up",
			action,
		)
	}
	return nil

}

func (m *maxTriesStrategy) Wait(_ error) interface{} {
	return nil
}

func (m *maxTriesStrategy) OnWaitExpired(_ WrapFunc, _ error, _ string) error {
	return nil
}
