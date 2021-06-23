package db_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/win5do/golang-microservice-demo/pkg/repository/db/dbcore"
)

func TestLock(t *testing.T) {
	i := 3
	wg := &sync.WaitGroup{}
	wg.Add(i)

	for i > 0 {
		holder := strconv.Itoa(i)
		action := "test"

		i--
		go func() {
			defer wg.Done()

			locker := dbcore.NewLockDb(action, holder, 10*time.Second)

			if _, err := locker.Lock(); err != nil {
				t.Logf("not hold the lock, err: %+v", err)
				return
			}

			time.Sleep(30 * time.Second)
			_ = locker.UnLock()
		}()
	}

	wg.Wait()
}
