package main

// contextError adds an operation to a failure while preserving errors.Is/As.
type contextError struct {
	operation string
	err       error
}

func (e *contextError) Error() string { return e.operation + ": " + e.err.Error() }
func (e *contextError) Unwrap() error { return e.err }

func wrapError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return &contextError{operation, err}
}
