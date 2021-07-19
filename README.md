# Retrier for Go

This library provides retries for functions in Go with configurable retry strategies.

## Installing

You can install this library using go modules:

```
go get github.com/haveyoudebuggedit/go-retry
```

## Basic usage

The simplest usage is calling `retry.Retry()` as follows:

```go
if err := retry.Retry(
    "creating foo",
    func createFoo() error {
        return fmt.Errorf("creating foo failed")
    },
); err != nil {
    panic(err)
}
```

This will retry the `createFoo()` function maximum 30 times.

## OOP-style call

You can also call the Retry function using the `New()` method:

```go
retrier := retry.New(
    logger ,
    wrapFunc,
    defaultClassifiers,
    defaultBackoff,
    defaultTimeout,
)
```

The **logger** function is optional and can be a function that fits the following signature: `func(format string, args ...interface{})`.

The **wrapFunc** parameter also optional and contains a function used to create a wrapped error. The function signature is the following: `func(err error, format string, args ...interface{}) error`

The remaining parameters are default strategies for retrying. This will be explained below.

Once you have the `retrier` you can call the `Retry` function just like the procedural counterpart.

## Customizing behavior

You can customize the retry behavior by passing retry strategies to the `retry.Retry` function:

```go
retry.Retry(
    action,
    call,
    strategy1,
    strategy2,
    strategy3
)
```

Below we'll explain the default strategies that are included.

### Exponential backoff strategy

The exponential backoff strategy adds an increasing delay between retries, starting at 1 second. The factor is customizable.

You can create this strategy by calling `retry.ExponentialBackoff(factor)`.

### Context strategy

You can cancel a retry when a Go [context](https://pkg.go.dev/context) expires. This strategy can be created by calling `retry.Context(ctx)`.

### Timeout strategy

You can cancel a retry when a timer expires. This timer starts counting when you **create** the strategy.

This is useful when passing the strategy to multiple API calls in sequence and you want an overall timeout.

You can create this strategy wth the duration parameter by calling `retry.Timeout(duration)`.

### Call Timeout strategy

You can also cancel a retry when a timer expires that is started when the strategy is used.

This is useful when you want to pass te strategy to multiple API calls and want an individual timeout for each API call individually.

You can create this strategy by calling `retry.CallTimeout(duration)`

### Maximum tries strategy

You can cancel retries when a maximum number of tries has been reached. You can create this strategy by calling `retry.MaxTries(tries)`.

## Writing a custom strategy

You can also write a custom strategy. First, you have to set up a factory function for your strategy. For this, you need to know the features of your strategy:

- `canClassify` means it can decide if the strategy can classify errors if they should or shouldn't be retried.
- `canWait` means the strategy can return a channel that is closed when a timer expires.
- `canTimeout` means the strategy can decide if no more retries should be attempted.

These values are used to add default strategies as described above.

```go
func MyStrategy() retry.Strategy {
    canClassify := true
    canWait := false
    canTimeout := false
    return retry.NewStrategy(
        func() retry.StrategyInstance {
            return &yourStrategyInstance{}        
        },
        canClassify,
        canWait,
        canTimeout,
    )
}
```

The strategy instance must implement the following interface:

```go
type StrategyInstance interface {
    // CanRetry returns an error if no more tries should be attempted.
    // The error will be returned directly from the retry function. The
    // passed action parameters can be used to create a meaningful error
    // message.
    CanRetry(wrap WrapFunc, err error, action string) error
    // Wait returns a channel that is closed when the wait time expires. The
    // channel can have any content, so it is provided as an interface{}.
    // This function may return nil if it doesn't provide a wait time.
    Wait(err error) interface{}
    // OnWaitExpired is a hook that gives the strategy the option to return
    // an error if its wait has expired. It will only be called if it is the
    // first to reach its wait. If no error is returned the loop is
    // continued. The passed action name can be incorporated into an error
    // message.
    OnWaitExpired(wrap WrapFunc, err error, action string) error
}
```