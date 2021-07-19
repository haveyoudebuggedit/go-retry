package retry

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

// Retry is the Retrier configured with sensible defaults. The defaults added are:
//
// - Exponential backoff with a factor of 2.
// - Maximum tries set to 30.
var Retry = New(
	log.Printf,
	nil,
	nil,
	[]Strategy{ExponentialBackoff(2)},
	[]Strategy{MaxTries(30)},
)

// New creates a new retrier with the specified parameters.
//
// - logger is a function that takes log messages. Can be set to nil to disable logging.
// - wrapFunc is a function to create wrapped errors.
// - defaultClassifiers is a list of strategies that
func New(
	logger func(format string, args ...interface{}),
	wrapFunc WrapFunc,
	defaultClassifiers []Strategy,
	defaultBackoff []Strategy,
	defaultTimeout []Strategy,
) Retrier {
	if logger == nil {
		logger = func(_ string, _ ...interface{}) {}
	}
	if wrapFunc == nil {
		wrapFunc = func(err error, format string, args ...interface{}) error {
			return fmt.Errorf(fmt.Sprintf("%s (%s)", format, "%w"), append(args, err))
		}
	}
	return &retrier{
		logger:             logger,
		wrapFunc:           wrapFunc,
		defaultClassifiers: defaultClassifiers,
		defaultBackoff:     defaultBackoff,
		defaultTimeout:     defaultTimeout,
	}
}

// Retrier is an interface that holds a retry function.
type Retrier interface {
	// Retry performs a number of retries according to the passed strategies.
	// The what parameter is the name of the action performed in the "ing" form. (e.g. creating x)
	// The call parameter contains the function to retry.
	Retry(
		what string,
		call func() error,
		strategies ...Strategy,
	) error
}

type retrier struct {
	logger             func(format string, args ...interface{})
	wrapFunc           WrapFunc
	defaultClassifiers []Strategy
	defaultBackoff     []Strategy
	defaultTimeout     []Strategy
}

func (r *retrier) Retry(action string, call func() error, strategies ...Strategy) error {
	strategies = defaultRetries(strategies, r.defaultBackoff, r.defaultTimeout, r.defaultClassifiers)

	retries := make([]StrategyInstance, len(strategies))
	for i, factory := range strategies {
		retries[i] = factory.Get()
	}

	r.logger("%s%s...", strings.ToUpper(action[:1]), action[1:])
	for {
		err := call()
		if err == nil {
			r.logger("Completed %s.", action)
			return nil
		}
		for _, retry := range retries {
			if err := retry.CanRetry(r.wrapFunc, err, action); err != nil {
				r.logger("Error while %s. (%v)", action, err)
				return err
			}
		}
		var chans []reflect.SelectCase
		for _, r := range retries {
			c := r.Wait(err)
			if c != nil {
				chans = append(chans, reflect.SelectCase{
					Dir:  reflect.SelectRecv,
					Chan: reflect.ValueOf(c),
					Send: reflect.Value{},
				})
			}
		}
		if len(chans) == 0 {
			r.logger(
				"No retry strategies with waiting function specified for %s.",
				action,
			)
			panic(fmt.Errorf("no retry strategies with waiting function specified for %s", action))
		}
		chosen, _, _ := reflect.Select(chans)
		if err := retries[chosen].OnWaitExpired(r.wrapFunc, err, action); err != nil {
			r.logger("Error while %s. (%v)", action, err)
			return err
		}
	}
}

func defaultRetries(retries []Strategy, backoff []Strategy, timeout []Strategy, classifier []Strategy) []Strategy {
	foundWait := false
	foundTimeout := false
	foundClassifier := false
	for _, r := range retries {
		if r.CanWait() {
			foundWait = true
		}
		if r.CanTimeout() {
			foundTimeout = true
		}
		if r.CanClassifyErrors() {
			foundClassifier = true
		}
	}
	if !foundWait {
		retries = append(retries, backoff...)
	}
	if !foundTimeout {
		retries = append(retries, timeout...)
	}
	if !foundClassifier {
		retries = append(retries, classifier...)
	}
	return retries
}
