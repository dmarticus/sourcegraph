package errors

import (
	"context"
)

// Ignore filters out any errors that match pred. This applies
// recursively to MultiErrors, filtering out any child errors
// that match `pred`, or returning `nil` if all of the child
// errors match `pred`.
func Ignore(err error, pred ErrorPredicate) error {
	// If the error (or any wrapped error) is a multierror,
	// filter its children.
	var multi *MultiError
	if As(err, &multi) {
		filtered := multi.Errors[:0]
		for _, childErr := range multi.Errors {
			if ignored := Ignore(childErr, pred); ignored != nil {
				filtered = append(filtered, ignored)
			}
		}
		if len(filtered) == 0 {
			return nil
		}
		multi.Errors = filtered
		return err
	}

	if pred(err) {
		return nil
	}
	return err
}

// ErrorPredicate is a function type that returns whether an error matches a given condition
type ErrorPredicate func(error) bool

// HasTypePred returns an ErrorPredicate that returns true for errors that unwrap to an error with the same type as target
func HasTypePred(target error) ErrorPredicate {
	return func(err error) bool {
		return HasType(err, target)
	}
}

// IsPred returns an ErrorPredicate that returns true for errors that uwrap to the target error
func IsPred(target error) ErrorPredicate {
	return func(err error) bool {
		return Is(err, target)
	}
}

var IsContextCanceled = IsPred(context.Canceled)
