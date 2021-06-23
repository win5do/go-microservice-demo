package util

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"
)

func GetEnvOrDefault(key string, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	} else {
		return def
	}
}

func ToBool(val string) bool {
	return strings.ToLower(val) == "true"
}

func ToIntOrDie(val string) int {
	r, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}
	return r
}

type ctxKeyWaitGroup struct{}

func GetWaitGroupInCtx(ctx context.Context) *sync.WaitGroup {
	if wg, ok := ctx.Value(ctxKeyWaitGroup{}).(*sync.WaitGroup); ok {
		return wg
	}

	return nil
}

func NewWaitGroupCtx() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.WithValue(context.Background(), ctxKeyWaitGroup{}, new(sync.WaitGroup)))
}
