package errx

import (
	"errors"
	"testing"

	errors2 "github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func errOnce() error {
	e1 := errors.New("normal error")
	e2 := WithStackOnce(e1)
	e3 := WithStackOnce(e2)
	return WithStackOnce(e3)
}

func TestWithStackOnce(t *testing.T) {
	err := errOnce()
	serr := err.(stackTracer)
	t.Logf("err: %+v", err)
	require.Condition(t, func() bool {
		return len(serr.StackTrace()) < 6
	})
}

func errMulti() error {
	e1 := errors.New("normal error")
	e2 := errors2.WithStack(e1)
	e3 := errors2.WithStack(e2)
	return errors2.WithStack(e3)
}

func TestOriginWithStack(t *testing.T) {
	t.Logf("err: %+v", errMulti())
}
