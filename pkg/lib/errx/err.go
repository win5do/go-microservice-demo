package errx

import (
	errors2 "github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors2.StackTrace
}

var stackFlag = true

func SetFlag(b bool) {
	stackFlag = b
}

func WithStackOnce(err error) error {
	if !stackFlag {
		return err
	}

	_, ok := err.(stackTracer)
	if ok {
		return err
	}

	return errors2.WithStack(err)
}
