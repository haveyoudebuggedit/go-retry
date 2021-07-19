package retry

// Strategy is a plugin to the Retrier that provides a hook to manipulate retry behavior.
type Strategy interface {
	// Get returns the retry strategy instance itself. The instance may have internal state to keep track of things like
	// how much time has elapsed since Get() was called.
	Get() StrategyInstance

	// CanClassifyErrors indicates if the strategy can determine if an error is retryable. At least one strategy with
	// this capability needs to be passed.
	CanClassifyErrors() bool
	// CanWait indicates if the retry strategy can wait in a loop. At least one strategy with this capability
	// needs to be passed.
	CanWait() bool
	// CanTimeout indicates that the retry strategy can properly abort a loop. At least one retry strategy with
	// this capability needs to be passed.
	CanTimeout() bool
}

// StrategyInstance is the instantiation of a strategy. It may have internal state to keep track of things like how much
// time has been elapsed since it was created.
type StrategyInstance interface {
	// CanRetry returns an error if no more tries should be attempted. The error will be returned directly from the
	// retry function. The passed action parameters can be used to create a meaningful error message.
	CanRetry(wrap WrapFunc, err error, action string) error
	// Wait returns a channel that is closed when the wait time expires. The channel can have any content, so it is
	// provided as an interface{}. This function may return nil if it doesn't provide a wait time.
	Wait(err error) interface{}
	// OnWaitExpired is a hook that gives the strategy the option to return an error if its wait has expired. It will
	// only be called if it is the first to reach its wait. If no error is returned the loop is continued. The passed
	// action name can be incorporated into an error message.
	OnWaitExpired(wrap WrapFunc, err error, action string) error
}

// NewStrategy creates a wrapper for a strategy instance with the specified parameters.
func NewStrategy(
	factory func() StrategyInstance,
	canClassifyErrors bool,
	canWait bool,
	canTimeout bool,
) Strategy {
	return &strategy{
		factory:           factory,
		canClassifyErrors: canClassifyErrors,
		canWait:           canWait,
		canTimeout:        canTimeout,
	}
}

type strategy struct {
	factory           func() StrategyInstance
	canClassifyErrors bool
	canWait           bool
	canTimeout        bool
}

func (s *strategy) Get() StrategyInstance {
	return s.factory()
}

func (s *strategy) CanClassifyErrors() bool {
	return s.canClassifyErrors
}

func (s *strategy) CanWait() bool {
	return s.canWait
}

func (s *strategy) CanTimeout() bool {
	return s.canTimeout
}
