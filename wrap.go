package retry

// WrapFunc is a function that can wrap an error and format an error message. This is provided
// so that the error generation can be customized.
type WrapFunc func(err error, format string, args ...interface{}) error
